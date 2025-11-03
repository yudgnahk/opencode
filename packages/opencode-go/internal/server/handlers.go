package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/opencode/opencode-go/internal/storage"
)

type Server struct {
	storage *storage.Storage
}

func New(storage *storage.Storage) *Server {
	return &Server{
		storage: storage,
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

// Project stubs
func (s *Server) handleProjectCurrent(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": "stub"})
}

func (s *Server) handleProjectInit(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.WriteHeader(http.StatusCreated)
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

// File operation stubs
func (s *Server) handleFilesRead(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": ""})
}

func (s *Server) handleFilesWrite(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleFilesEdit(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleFilesGlob(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]string{})
}

func (s *Server) handleFilesGrep(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

// Bash stub
func (s *Server) handleBashExecute(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement in Phase 2
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"output": ""})
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
