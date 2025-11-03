package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const serverURL = "http://localhost:3000"

type Session struct {
	ID string `json:"id"`
}

type Message struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type Event struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId"`
	Data      json.RawMessage `json:"data"`
}

func createSession() (string, error) {
	body := map[string]string{
		"projectId": "test-sse",
		"provider":  "anthropic",
		"model":     "claude-3-5-sonnet-20241022",
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(serverURL+"/api/sessions", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return "", err
	}

	return session.ID, nil
}

func addMessage(sessionID, text string) error {
	body := map[string]any{
		"role": "user",
		"content": []map[string]string{
			{"type": "text", "text": text},
		},
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := http.Post(serverURL+"/api/sessions/"+sessionID+"/messages", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func streamEvents(ctx context.Context, sessionID string, clientID int, wg *sync.WaitGroup) {
	defer wg.Done()

	req, _ := http.NewRequestWithContext(ctx, "GET", serverURL+"/api/sessions/"+sessionID+"/stream", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Client %d: Error connecting to stream: %v\n", clientID, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Client %d: Connected to SSE stream\n", clientID)

	scanner := bufio.NewScanner(resp.Body)
	eventCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		// Parse SSE format
		if len(line) > 6 && line[:6] == "event:" {
			eventType := line[7:]
			fmt.Printf("Client %d: Received event type: %s\n", clientID, eventType)
		} else if len(line) > 5 && line[:5] == "data:" {
			eventCount++
			fmt.Printf("Client %d: Received event #%d\n", clientID, eventCount)
		} else if line[:1] == ":" {
			// Keepalive comment
			fmt.Printf("Client %d: Keepalive\n", clientID)
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		fmt.Printf("Client %d: Scanner error: %v\n", clientID, err)
	}

	fmt.Printf("Client %d: Disconnected (received %d events)\n", clientID, eventCount)
}

func main() {
	fmt.Println("SSE Streaming Test Client")
	fmt.Println("==========================")

	// Create a session
	fmt.Println("\n1. Creating session...")
	sessionID, err := createSession()
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}
	fmt.Printf("✓ Session created: %s\n", sessionID)

	// Start multiple SSE clients
	fmt.Println("\n2. Starting 3 SSE clients...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go streamEvents(ctx, sessionID, i, &wg)
		time.Sleep(100 * time.Millisecond) // Stagger connections
	}

	// Give clients time to connect
	time.Sleep(500 * time.Millisecond)

	// Add messages while clients are connected
	fmt.Println("\n3. Adding messages...")
	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("Test message #%d", i)
		if err := addMessage(sessionID, msg); err != nil {
			fmt.Printf("Error adding message: %v\n", err)
		} else {
			fmt.Printf("✓ Added message: %s\n", msg)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for timeout or all clients to finish
	fmt.Println("\n4. Waiting for clients to complete...")
	wg.Wait()

	fmt.Println("\n✓ Test completed successfully!")
	fmt.Println("\nExpected result: Each client should receive 5 message_added events")
}
