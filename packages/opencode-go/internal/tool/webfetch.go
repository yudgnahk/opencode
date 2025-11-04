package tool

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

type WebFetchRequest struct {
	URL     string `json:"url"`
	Format  string `json:"format"`  // "text", "markdown", "html"
	Timeout int    `json:"timeout"` // seconds
}

type WebFetchResponse struct {
	Content     string            `json:"content"`
	ContentType string            `json:"contentType"`
	StatusCode  int               `json:"statusCode"`
	Headers     map[string]string `json:"headers"`
}

func (t *ToolExecutor) WebFetch(req WebFetchRequest) (*WebFetchResponse, error) {
	if req.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	// Upgrade HTTP to HTTPS automatically (except for localhost/127.0.0.1)
	if strings.HasPrefix(req.URL, "http://") {
		// Don't upgrade localhost or 127.0.0.1 (for testing and local development)
		if !strings.Contains(req.URL, "localhost") && !strings.Contains(req.URL, "127.0.0.1") {
			req.URL = "https://" + strings.TrimPrefix(req.URL, "http://")
		}
	}

	// Default timeout: 30 seconds
	timeout := 30
	if req.Timeout > 0 {
		timeout = req.Timeout
	}
	if timeout > 120 {
		timeout = 120 // Max 2 minutes
	}

	// Default format
	if req.Format == "" {
		req.Format = "markdown"
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("User-Agent", "OpenCode/1.0")

	// Make request
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	content := string(body)
	contentType := resp.Header.Get("Content-Type")

	// Convert to requested format
	if req.Format == "markdown" && isHTML(contentType) {
		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(content)
		if err == nil {
			content = markdown
		}
		// If conversion fails, keep original HTML
	} else if req.Format == "text" && isHTML(contentType) {
		// For text format on HTML, also convert to markdown first (which strips tags)
		// then we could further simplify, but markdown is clean enough
		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(content)
		if err == nil {
			content = markdown
		}
	}

	// Extract headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return &WebFetchResponse{
		Content:     content,
		ContentType: contentType,
		StatusCode:  resp.StatusCode,
		Headers:     headers,
	}, nil
}

// isHTML checks if content type indicates HTML
func isHTML(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml")
}
