package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/opencode/opencode-go/internal/project"
	"github.com/opencode/opencode-go/internal/storage"
	"github.com/opencode/opencode-go/internal/tool"
)

type Server struct {
	storage *storage.Storage
	tools   *tool.ToolExecutor
	project *project.Project
}

func New(storage *storage.Storage, workingDir string) *Server {
	// Detect project
	proj, _ := project.Detect(workingDir)

	return &Server{
		storage: storage,
		tools:   tool.NewToolExecutor(workingDir),
		project: proj,
	}
}

func (s *Server) RegisterRoutes(r chi.Router) {
	// Health & Status
	r.Get("/health", s.handleHealth)
	r.Get("/api/status", s.handleStatus)

	// Project
	r.Get("/api/project/current", s.handleProjectCurrent)
	r.Post("/api/project/init", s.handleProjectInit)

	// Sessions
	r.Get("/api/sessions", s.handleSessionsList)
	r.Post("/api/sessions", s.handleSessionsCreate)
	r.Get("/api/sessions/{id}", s.handleSessionsGet)
	r.Delete("/api/sessions/{id}", s.handleSessionsDelete)
	r.Post("/api/sessions/{id}/fork", s.handleSessionsFork)

	// Messages
	r.Get("/api/sessions/{id}/messages", s.handleMessagesGet)
	r.Post("/api/sessions/{id}/messages", s.handleMessagesCreate)
	r.Get("/api/sessions/{id}/stream", s.handleMessagesStream)

	// Agents
	r.Get("/api/agents", s.handleAgentsList)
	r.Get("/api/agents/{id}", s.handleAgentsGet)

	// Tools
	r.Get("/api/tools", s.handleToolsList)
	r.Post("/api/tools/{name}/execute", s.handleToolsExecute)

	// Configuration
	r.Get("/api/config", s.handleConfigGet)
	r.Post("/api/config", s.handleConfigUpdate)

	// File operations
	r.Post("/api/files/read", s.handleFilesRead)
	r.Post("/api/files/write", s.handleFilesWrite)
	r.Post("/api/files/edit", s.handleFilesEdit)
	r.Post("/api/files/glob", s.handleFilesGlob)
	r.Post("/api/files/grep", s.handleFilesGrep)

	// Bash
	r.Post("/api/bash/execute", s.handleBashExecute)

	// LSP
	r.Post("/api/lsp/diagnostics", s.handleLSPDiagnostics)
	r.Post("/api/lsp/hover", s.handleLSPHover)

	// Auth
	r.Post("/api/auth/login", s.handleAuthLogin)
	r.Post("/api/auth/logout", s.handleAuthLogout)

	// Todos
	r.Get("/api/sessions/{id}/todos", s.handleTodosGet)
	r.Post("/api/sessions/{id}/todos", s.handleTodosUpdate)
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

// Session stubs
func (s *Server) handleSessionsList(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 3
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleSessionsCreate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 3
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": "stub-session"})
}

func (s *Server) handleSessionsGet(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 3
	sessionID := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": sessionID})
}

func (s *Server) handleSessionsDelete(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 3
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSessionsFork(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 3
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": "stub-fork"})
}

// Message stubs
func (s *Server) handleMessagesGet(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 3
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleMessagesCreate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 4
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": "stub-message"})
}

func (s *Server) handleMessagesStream(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 4 (SSE streaming)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
}

// Agent stubs
func (s *Server) handleAgentsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleAgentsGet(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": agentID})
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

// LSP stubs
func (s *Server) handleLSPDiagnostics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 5
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleLSPHover(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 5
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{})
}

// Auth stubs
func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": "stub-token"})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Todo stubs
func (s *Server) handleTodosGet(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 5
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleTodosUpdate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 5
	w.WriteHeader(http.StatusOK)
}
