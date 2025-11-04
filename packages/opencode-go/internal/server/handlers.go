package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opencode/opencode-go/internal/auth"
	"github.com/opencode/opencode-go/internal/config"
	"github.com/opencode/opencode-go/internal/project"
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/session"
	"github.com/opencode/opencode-go/internal/storage"
	"github.com/opencode/opencode-go/internal/tool"
)

type Server struct {
	storage           *storage.Storage
	tools             *tool.ToolExecutor
	taskExecutor      *tool.TaskExecutor
	project           *project.Project
	sessionManager    *session.Manager
	providerRegistry  *provider.Registry
	completionService *session.CompletionService
	workingDir        string
}

func New(storage *storage.Storage, workingDir string) *Server {
	// Detect project
	proj, _ := project.Detect(workingDir)

	// Create provider registry
	registry := provider.NewRegistry()

	// Create session manager
	sessionMgr := session.NewManager(storage, registry)

	// Create completion service
	completionSvc := session.NewCompletionService(sessionMgr, registry)

	// Determine project ID
	projectID := "default"
	if proj != nil {
		projectID = proj.Name
	}

	return &Server{
		storage:           storage,
		tools:             tool.NewToolExecutor(workingDir, storage, registry, projectID),
		taskExecutor:      tool.NewTaskExecutor(storage, registry, projectID),
		project:           proj,
		sessionManager:    sessionMgr,
		providerRegistry:  registry,
		completionService: completionSvc,
		workingDir:        workingDir,
	}
}

func (s *Server) InitProviders(ctx context.Context) error {
	retryConfig := provider.DefaultRetryConfig()

	// Helper to get API key from auth system or env var
	getAPIKey := func(providerID, envVar string) string {
		// Check auth system first
		authInfo, err := auth.Get(providerID)
		if err == nil && authInfo != nil {
			if apiAuth, ok := authInfo.(auth.API); ok && apiAuth.Key != "" {
				return apiAuth.Key
			}
		}
		// Fall back to environment variable
		return os.Getenv(envVar)
	}

	// Helper to register or update a provider
	registerProvider := func(p provider.Provider) error {
		name := p.Name()
		// Unregister if already exists
		s.providerRegistry.Unregister(name)
		// Register the new/updated provider
		return s.providerRegistry.Register(p)
	}

	// Initialize Anthropic provider with retry
	if apiKey := getAPIKey("anthropic", "ANTHROPIC_API_KEY"); apiKey != "" {
		anthropic := provider.NewAnthropic(apiKey)
		wrapped := provider.NewRetryableProvider(anthropic, retryConfig)
		if err := registerProvider(wrapped); err != nil {
			return fmt.Errorf("failed to register anthropic: %w", err)
		}
	}

	// Initialize OpenAI provider with retry
	if apiKey := getAPIKey("openai", "OPENAI_API_KEY"); apiKey != "" {
		openai := provider.NewOpenAI(apiKey)
		wrapped := provider.NewRetryableProvider(openai, retryConfig)
		if err := registerProvider(wrapped); err != nil {
			return fmt.Errorf("failed to register openai: %w", err)
		}
	}

	// Initialize Gemini provider with retry
	if apiKey := getAPIKey("gemini", "GEMINI_API_KEY"); apiKey != "" {
		gemini, err := provider.NewGemini(ctx, apiKey)
		if err != nil {
			return fmt.Errorf("failed to create gemini provider: %w", err)
		}
		wrapped := provider.NewRetryableProvider(gemini, retryConfig)
		if err := registerProvider(wrapped); err != nil {
			return fmt.Errorf("failed to register gemini: %w", err)
		}
	}

	return nil
}

func (s *Server) RegisterRoutes(r chi.Router) {
	// Health & Status
	r.Get("/health", s.handleHealth)
	r.Get("/status", s.handleStatus)

	// Path
	r.Get("/path", s.handlePath)

	// Project
	r.Get("/project/current", s.handleProjectCurrent)
	r.Post("/project/init", s.handleProjectInit)

	// Sessions
	r.Get("/session", s.handleSessionsList)
	r.Post("/session", s.handleSessionsCreate)
	r.Get("/session/{id}", s.handleSessionsGet)
	r.Delete("/session/{id}", s.handleSessionsDelete)
	r.Post("/session/{id}/fork", s.handleSessionsFork)
	r.Post("/session/{id}/branch", s.handleSessionsBranch)
	r.Post("/session/{id}/revert", s.handleSessionsRevert)
	r.Post("/session/{id}/summarize", s.handleSessionsSummarize)

	// Todos
	r.Get("/session/{id}/todos", s.handleTodosGet)
	r.Put("/session/{id}/todos", s.handleTodosUpdate)
	r.Post("/session/{id}/todos", s.handleTodosAdd)
	r.Patch("/session/{id}/todos/{todoId}", s.handleTodosUpdateStatus)

	// Tasks
	r.Post("/session/{id}/task", s.handleTaskExecute)

	// Messages
	r.Get("/session/{id}/messages", s.handleMessagesGet)
	r.Post("/session/{id}/messages", s.handleMessagesCreate)
	r.Get("/session/{id}/stream", s.handleMessagesStream)

	// AI Completions
	r.Post("/session/{id}/complete", s.handleAIComplete)
	r.Post("/session/{id}/complete/stream", s.handleAICompleteStream)

	// Agents
	r.Get("/agent", s.handleAgentsList)
	r.Get("/agent/{id}", s.handleAgentsGet)

	// Commands
	r.Get("/command", s.handleCommandsList)

	// Tools
	r.Get("/tool", s.handleToolsList)
	r.Post("/tool/{name}/execute", s.handleToolsExecute)

	// Configuration
	r.Get("/config", s.handleConfigGet)
	r.Post("/config", s.handleConfigUpdate)

	// File operations
	r.Get("/file/status", s.handleFilesStatus)
	r.Post("/file/read", s.handleFilesRead)
	r.Post("/file/write", s.handleFilesWrite)
	r.Post("/file/edit", s.handleFilesEdit)
	r.Post("/file/glob", s.handleFilesGlob)
	r.Post("/file/grep", s.handleFilesGrep)

	// Bash
	r.Post("/bash/execute", s.handleBashExecute)

	// LSP
	r.Post("/lsp/hover", s.handleLSPHover)
	r.Post("/lsp/completion", s.handleLSPCompletion)
	r.Post("/lsp/definition", s.handleLSPDefinition)
	r.Post("/lsp/references", s.handleLSPReferences)

	// Logging
	r.Post("/log", s.handleLog)

	// Events
	r.Get("/event", s.handleEvent)

	// TUI Control
	r.Get("/tui/control/next", s.handleTuiControlNext)
	r.Post("/tui/control/response", s.handleTuiControlResponse)

	// Auth
	r.Get("/auth", s.handleAuthList)
	r.Get("/auth/{provider}", s.handleAuthGet)
	r.Put("/auth/{provider}", s.handleAuthSet)
	r.Delete("/auth/{provider}", s.handleAuthDelete)

	// Todos
	r.Get("/session/{id}/todo", s.handleTodosGet)
	r.Post("/session/{id}/todo", s.handleTodosUpdate)
}

// Stub implementations (return empty responses)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version": "0.1.0-go",
		"backend": "go",
	})
}

func (s *Server) handlePath(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"directory": s.workingDir,
	}

	// Add state and config paths
	response["state"] = config.GetStatePath()
	response["config"] = config.GetConfigPath()

	// Add worktree path if git is detected
	if s.project != nil && s.project.Git != nil {
		response["worktree"] = s.project.Git.Root
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Project handlers
func (s *Server) handleProjectCurrent(w http.ResponseWriter, r *http.Request) {
	if s.project == nil {
		http.Error(w, "No project detected", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.project)
}

func (s *Server) handleProjectInit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	proj, err := project.Detect(req.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.project = proj

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(proj)
}

// Session handlers
func (s *Server) handleSessionsList(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("projectId")

	// Default to current project if not specified (matches TypeScript behavior)
	if projectID == "" && s.project != nil {
		projectID = s.project.ID
	}

	sessions, err := s.sessionManager.List(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (s *Server) handleSessionsCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID string `json:"projectId"`
		Provider  string `json:"provider"`
		Model     string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, err := s.sessionManager.Create(r.Context(), req.ProjectID, req.Provider, req.Model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (s *Server) handleSessionsGet(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	session, err := s.sessionManager.Get(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) handleSessionsDelete(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	if err := s.sessionManager.Delete(r.Context(), sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSessionsFork(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		FromMessageID string `json:"fromMessageId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	forked, err := s.sessionManager.Fork(r.Context(), sessionID, req.FromMessageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(forked)
}

func (s *Server) handleSessionsBranch(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		FromMessageID string `json:"fromMessageId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	branched, err := s.sessionManager.Branch(r.Context(), sessionID, req.FromMessageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(branched)
}

func (s *Server) handleSessionsRevert(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		ToMessageID string `json:"toMessageId"`
		Inclusive   bool   `json:"inclusive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := s.sessionManager.Revert(r.Context(), sessionID, req.ToMessageID, req.Inclusive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSessionsSummarize(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	summary, err := s.sessionManager.Summarize(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Summary string `json:"summary"`
	}{
		Summary: summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Todo handlers
func (s *Server) handleTodosGet(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	todos, err := s.sessionManager.GetTodos(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func (s *Server) handleTodosUpdate(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		Todos []session.Todo `json:"todos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.sessionManager.UpdateTodos(r.Context(), sessionID, req.Todos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleTodosAdd(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		Content  string `json:"content"`
		Priority string `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.sessionManager.AddTodo(r.Context(), sessionID, req.Content, req.Priority); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleTodosUpdateStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	todoID := chi.URLParam(r, "todoId")

	var req struct {
		Status session.TodoStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.sessionManager.UpdateTodoStatus(r.Context(), sessionID, todoID, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Task handlers
func (s *Server) handleTaskExecute(w http.ResponseWriter, r *http.Request) {
	var req tool.TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.taskExecutor.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Message handlers
func (s *Server) handleMessagesGet(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	messages, err := s.sessionManager.GetMessages(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (s *Server) handleMessagesCreate(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		Role    session.Role           `json:"role"`
		Content []session.ContentBlock `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err := s.sessionManager.AddMessage(r.Context(), sessionID, req.Role, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func (s *Server) handleMessagesStream(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Subscribe to events
	eventCh := s.sessionManager.EventBus().Subscribe(sessionID)
	defer s.sessionManager.EventBus().Unsubscribe(sessionID, eventCh)

	// Send keepalive every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event := <-eventCh:
			// Marshal event to JSON
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			// Write SSE format
			fmt.Fprintf(w, "event: %s\n", event.Type)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

		case <-ticker.C:
			// Send keepalive
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// Agent stubs
func (s *Server) handleAgentsList(w http.ResponseWriter, r *http.Request) {
	// Return default built-in agents
	agents := []map[string]interface{}{
		{
			"name":        "general",
			"description": "General-purpose agent for researching complex questions, searching for code, and executing multi-step tasks.",
			"mode":        "subagent",
			"builtIn":     true,
			"permission": map[string]interface{}{
				"edit": "allow",
				"bash": map[string]string{
					"*": "allow",
				},
				"webfetch": "allow",
			},
			"tools":   map[string]bool{},
			"options": map[string]interface{}{},
		},
		{
			"name":    "build",
			"mode":    "primary",
			"builtIn": true,
			"permission": map[string]interface{}{
				"edit": "allow",
				"bash": map[string]string{
					"*": "allow",
				},
				"webfetch": "allow",
			},
			"tools":   map[string]bool{},
			"options": map[string]interface{}{},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (s *Server) handleAgentsGet(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": agentID})
}

// Command stubs
func (s *Server) handleCommandsList(w http.ResponseWriter, r *http.Request) {
	// Return empty array - commands are defined in config file
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

// Tool stubs
func (s *Server) handleToolsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleToolsExecute(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": "stub"})
}

// Config stubs
func (s *Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (s *Server) handleConfigUpdate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// File operation handlers
func (s *Server) handleFilesStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement git diff status
	// For now, return empty array
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleFilesRead(w http.ResponseWriter, r *http.Request) {
	var req tool.ReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.Read(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleFilesWrite(w http.ResponseWriter, r *http.Request) {
	var req tool.WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.Write(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleFilesEdit(w http.ResponseWriter, r *http.Request) {
	var req tool.EditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.Edit(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleFilesGlob(w http.ResponseWriter, r *http.Request) {
	var req tool.GlobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.Glob(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleFilesGrep(w http.ResponseWriter, r *http.Request) {
	var req tool.GrepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.Grep(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Bash handler
func (s *Server) handleBashExecute(w http.ResponseWriter, r *http.Request) {
	var req tool.BashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.Bash(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LSP handlers

func (s *Server) handleLSPHover(w http.ResponseWriter, r *http.Request) {
	var req tool.LSPHoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.LSPHover(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleLSPCompletion(w http.ResponseWriter, r *http.Request) {
	var req tool.LSPCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.LSPCompletion(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleLSPDefinition(w http.ResponseWriter, r *http.Request) {
	var req tool.LSPDefinitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.LSPDefinition(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleLSPReferences(w http.ResponseWriter, r *http.Request) {
	var req tool.LSPReferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := s.tools.LSPReferences(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Auth handlers

// handleAuthList returns all stored authentication providers
func (s *Server) handleAuthList(w http.ResponseWriter, r *http.Request) {
	allAuth, err := auth.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to a simple map with provider names (masking sensitive data)
	result := make(map[string]string)
	for key, info := range allAuth {
		result[key] = string(info.GetType())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleAuthGet retrieves auth info for a specific provider (masked)
func (s *Server) handleAuthGet(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")

	authInfo, err := auth.Get(providerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if authInfo == nil {
		http.Error(w, "provider not found", http.StatusNotFound)
		return
	}

	// Return masked info (just type, not sensitive data)
	result := map[string]string{
		"provider": providerID,
		"type":     string(authInfo.GetType()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleAuthSet stores authentication credentials for a provider
func (s *Server) handleAuthSet(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")

	// Parse the request body as raw JSON first
	var rawData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert back to JSON for parsing
	jsonData, err := json.Marshal(rawData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Determine type and parse accordingly
	typeStr, ok := rawData["type"].(string)
	if !ok {
		http.Error(w, "missing or invalid 'type' field", http.StatusBadRequest)
		return
	}

	var authInfo auth.Info
	switch auth.AuthType(typeStr) {
	case auth.TypeAPI:
		var api auth.API
		if err := json.Unmarshal(jsonData, &api); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		authInfo = api
	case auth.TypeOAuth:
		var oauth auth.OAuth
		if err := json.Unmarshal(jsonData, &oauth); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		authInfo = oauth
	case auth.TypeWellKnown:
		var wellknown auth.WellKnown
		if err := json.Unmarshal(jsonData, &wellknown); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		authInfo = wellknown
	default:
		http.Error(w, "invalid auth type", http.StatusBadRequest)
		return
	}

	// Store the auth info
	if err := auth.Set(providerID, authInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Re-initialize providers to pick up the new auth
	if err := s.InitProviders(r.Context()); err != nil {
		http.Error(w, fmt.Sprintf("failed to reinitialize providers: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleAuthDelete removes authentication credentials for a provider
func (s *Server) handleAuthDelete(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider")

	if err := auth.Remove(providerID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Log handler
func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement proper logging
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(true)
}

// Event handler
func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement event streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Keep connection open
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// TUI Control handlers
func (s *Server) handleTuiControlNext(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement TUI control queue
	// For now, just block to keep the connection open
	<-r.Context().Done()
}

func (s *Server) handleTuiControlResponse(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement TUI control response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(true)
}

// AI Completion handlers

func (s *Server) handleAIComplete(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Send message and get AI response
	if err := s.completionService.SendMessage(r.Context(), sessionID, req.Content); err != nil {
		http.Error(w, fmt.Sprintf("completion failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the latest message (the AI response)
	messages, err := s.sessionManager.GetMessages(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(messages) == 0 {
		http.Error(w, "no messages found", http.StatusInternalServerError)
		return
	}

	lastMessage := messages[len(messages)-1]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lastMessage)
}

func (s *Server) handleAICompleteStream(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Start streaming
	stream, err := s.completionService.StreamMessage(r.Context(), sessionID, req.Content)
	if err != nil {
		http.Error(w, fmt.Sprintf("streaming failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// Stream chunks to client
	for stream.Next() {
		chunk := stream.Chunk()

		// Convert chunk to JSON
		data, err := json.Marshal(chunk)
		if err != nil {
			continue
		}

		// Write SSE format
		fmt.Fprintf(w, "event: chunk\n")
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		// Stop on final chunk
		if chunk.Type == provider.ChunkTypeStop {
			break
		}
	}

	// Check for errors
	if err := stream.Error(); err != nil {
		errorData := map[string]string{"error": err.Error()}
		data, _ := json.Marshal(errorData)
		fmt.Fprintf(w, "event: error\n")
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
}
