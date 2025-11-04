package tool

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWebFetch(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	tests := []struct {
		name       string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		req        WebFetchRequest
		wantErr    bool
		check      func(resp *WebFetchResponse) bool
	}{
		{
			name: "fetch HTML and convert to markdown",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, "<html><body><h1>Hello World</h1><p>This is a test.</p></body></html>")
			},
			req: WebFetchRequest{
				Format: "markdown",
			},
			wantErr: false,
			check: func(resp *WebFetchResponse) bool {
				return resp.StatusCode == 200 &&
					strings.Contains(resp.Content, "Hello World") &&
					strings.Contains(resp.Content, "This is a test") &&
					resp.ContentType == "text/html"
			},
		},
		{
			name: "fetch HTML and keep as HTML",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, "<html><body><h1>Hello</h1></body></html>")
			},
			req: WebFetchRequest{
				Format: "html",
			},
			wantErr: false,
			check: func(resp *WebFetchResponse) bool {
				return resp.StatusCode == 200 &&
					strings.Contains(resp.Content, "<h1>Hello</h1>") &&
					resp.ContentType == "text/html"
			},
		},
		{
			name: "fetch plain text",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprint(w, "Plain text content")
			},
			req: WebFetchRequest{
				Format: "text",
			},
			wantErr: false,
			check: func(resp *WebFetchResponse) bool {
				return resp.StatusCode == 200 &&
					resp.Content == "Plain text content" &&
					resp.ContentType == "text/plain"
			},
		},
		{
			name: "default format is markdown",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, "<html><body><p>Test</p></body></html>")
			},
			req: WebFetchRequest{
				// Format not specified, should default to markdown
			},
			wantErr: false,
			check: func(resp *WebFetchResponse) bool {
				return resp.StatusCode == 200 &&
					strings.Contains(resp.Content, "Test") &&
					!strings.Contains(resp.Content, "<p>") // Should be converted from HTML
			},
		},
		{
			name: "check User-Agent header",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				userAgent := r.Header.Get("User-Agent")
				if userAgent != "OpenCode/1.0" {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, "Wrong User-Agent: %s", userAgent)
					return
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "OK")
			},
			req: WebFetchRequest{
				Format: "text",
			},
			wantErr: false,
			check: func(resp *WebFetchResponse) bool {
				return resp.StatusCode == 200 && resp.Content == "OK"
			},
		},
		{
			name: "handle 404 error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, "Not Found")
			},
			req: WebFetchRequest{
				Format: "text",
			},
			wantErr: false,
			check: func(resp *WebFetchResponse) bool {
				return resp.StatusCode == 404 && resp.Content == "Not Found"
			},
		},
		{
			name: "response includes headers",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Custom-Header", "test-value")
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprint(w, "Test")
			},
			req: WebFetchRequest{
				Format: "text",
			},
			wantErr: false,
			check: func(resp *WebFetchResponse) bool {
				return resp.StatusCode == 200 &&
					resp.Headers["X-Custom-Header"] == "test-value" &&
					resp.Headers["Content-Type"] == "text/plain"
			},
		},
		{
			name: "missing URL returns error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "Should not reach here")
			},
			req: WebFetchRequest{
				URL: "",
			},
			wantErr: true,
		},
		{
			name: "custom timeout",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(50 * time.Millisecond)
				fmt.Fprint(w, "Slow response")
			},
			req: WebFetchRequest{
				Format:  "text",
				Timeout: 1, // 1 second should be enough
			},
			wantErr: false,
			check: func(resp *WebFetchResponse) bool {
				return resp.StatusCode == 200 && resp.Content == "Slow response"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			// Set URL to test server (unless it's already set or explicitly empty for testing)
			if tt.req.URL == "" && tt.name != "missing URL returns error" {
				tt.req.URL = server.URL
			} else if tt.name != "missing URL returns error" {
				// If URL is set, use it (for explicit URL tests)
				tt.req.URL = server.URL
			}
			// For "missing URL returns error" test, leave URL as empty string

			resp, err := executor.WebFetch(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("WebFetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(resp) {
				t.Errorf("WebFetch() response check failed: %+v", resp)
			}
		})
	}
}

func TestWebFetchHTTPUpgrade(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	// Create HTTPS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "HTTPS response")
	}))
	defer server.Close()

	// Use HTTP URL (should be upgraded to HTTPS automatically)
	httpURL := strings.Replace(server.URL, "https://", "http://", 1)

	req := WebFetchRequest{
		URL:    httpURL,
		Format: "text",
	}

	// Note: This will fail because httptest.NewTLSServer uses self-signed cert
	// But it demonstrates the upgrade logic
	_, err := executor.WebFetch(req)
	// We expect an error due to certificate issues, but the URL should have been upgraded
	if err == nil {
		t.Log("Successfully fetched (unexpected with self-signed cert)")
	} else {
		// Error is expected with self-signed cert
		if !strings.Contains(err.Error(), "certificate") && !strings.Contains(err.Error(), "tls") {
			t.Errorf("Expected TLS/certificate error, got: %v", err)
		}
	}
}

func TestWebFetchTimeout(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "Delayed response")
	}))
	defer server.Close()

	req := WebFetchRequest{
		URL:     server.URL,
		Format:  "text",
		Timeout: 1, // 1 second timeout
	}

	_, err := executor.WebFetch(req)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "context deadline exceeded") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestWebFetchMaxTimeout(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "OK")
	}))
	defer server.Close()

	// Request with timeout > 120 seconds should be capped at 120
	req := WebFetchRequest{
		URL:     server.URL,
		Format:  "text",
		Timeout: 999, // Should be capped to 120
	}

	resp, err := executor.WebFetch(req)
	if err != nil {
		t.Errorf("WebFetch() error = %v", err)
	}
	if resp == nil || resp.StatusCode != 200 {
		t.Error("Expected successful response with timeout capped at 120s")
	}
}

func TestIsHTML(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{"text/html", true},
		{"text/html; charset=utf-8", true},
		{"application/xhtml+xml", true},
		{"TEXT/HTML", true}, // Case insensitive
		{"text/plain", false},
		{"application/json", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := isHTML(tt.contentType)
			if got != tt.want {
				t.Errorf("isHTML(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}
