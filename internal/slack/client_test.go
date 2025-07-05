package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")
	if client.token != "test-token" {
		t.Errorf("expected token 'test-token', got: %s", client.token)
	}
}

func TestTestAuth(t *testing.T) {
	t.Run("should succeed with valid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("expected Authorization header 'Bearer test-token', got: %s", r.Header.Get("Authorization"))
			}

			response := map[string]interface{}{
				"ok": true,
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient("test-token")
		client.httpClient = server.Client()

		req, _ := http.NewRequest("GET", server.URL+"/api/auth.test", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		resp, err := client.httpClient.Do(req)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		defer resp.Body.Close()

		var response struct {
			OK bool `json:"ok"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if !response.OK {
			t.Error("expected OK to be true")
		}
	})

	t.Run("should fail with invalid token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"ok":    false,
				"error": "invalid_auth",
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &Client{
			token:      "invalid-token",
			httpClient: server.Client(),
		}

		req, _ := http.NewRequest("GET", server.URL+"/api/auth.test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		resp, err := client.httpClient.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var response struct {
			OK    bool   `json:"ok"`
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.OK {
			t.Error("expected OK to be false")
		}
		if response.Error != "invalid_auth" {
			t.Errorf("expected error 'invalid_auth', got: %s", response.Error)
		}
	})
}

func TestListChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/conversations.list" {
			t.Errorf("expected path '/api/conversations.list', got: %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"ok": true,
			"channels": []Channel{
				{ID: "C1234567890", Name: "general"},
				{ID: "C0987654321", Name: "random"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	req, _ := http.NewRequest("GET", server.URL+"/api/conversations.list", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		OK       bool      `json:"ok"`
		Channels []Channel `json:"channels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.OK {
		t.Error("expected OK to be true")
	}

	if len(response.Channels) != 2 {
		t.Errorf("expected 2 channels, got: %d", len(response.Channels))
	}

	if response.Channels[0].Name != "general" {
		t.Errorf("expected first channel name 'general', got: %s", response.Channels[0].Name)
	}
}

func TestSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/search.messages") {
			t.Errorf("expected path to start with '/api/search.messages', got: %s", r.URL.Path)
		}

		query := r.URL.Query().Get("query")
		if query != "test-query" {
			t.Errorf("expected query 'test-query', got: %s", query)
		}

		response := map[string]interface{}{
			"ok": true,
			"messages": map[string]interface{}{
				"total": 1,
				"matches": []map[string]interface{}{
					{
						"type": "message",
						"text": "This is a test message",
						"user": "U1234567890",
						"username": "testuser",
						"channel": map[string]interface{}{
							"id":   "C1234567890",
							"name": "general",
						},
						"ts":        "1234567890.123456",
						"permalink": "https://test.slack.com/archives/C1234567890/p1234567890123456",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	req, _ := http.NewRequest("GET", server.URL+"/api/search.messages?query=test-query", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		OK       bool         `json:"ok"`
		Messages SearchResult `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.OK {
		t.Error("expected OK to be true")
	}

	if response.Messages.Total != 1 {
		t.Errorf("expected total 1, got: %d", response.Messages.Total)
	}

	if len(response.Messages.Matches) != 1 {
		t.Errorf("expected 1 match, got: %d", len(response.Messages.Matches))
	}

	match := response.Messages.Matches[0]
	if match.Text != "This is a test message" {
		t.Errorf("expected text 'This is a test message', got: %s", match.Text)
	}
}

func TestPostMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat.postMessage" {
			t.Errorf("expected path '/api/chat.postMessage', got: %s", r.URL.Path)
		}

		var requestData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if requestData["channel"] != "general" {
			t.Errorf("expected channel 'general', got: %s", requestData["channel"])
		}
		if requestData["text"] != "Hello, world!" {
			t.Errorf("expected text 'Hello, world!', got: %s", requestData["text"])
		}

		response := map[string]interface{}{
			"ok": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		httpClient: server.Client(),
	}

	data := map[string]interface{}{
		"channel": "general",
		"text":    "Hello, world!",
	}

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}

	req, _ := http.NewRequest("POST", server.URL+"/api/chat.postMessage", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.OK {
		t.Error("expected OK to be true")
	}
}