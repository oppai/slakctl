package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewSlackOAuthClient(t *testing.T) {
	client := NewSlackOAuthClient("test-client-id", "test-client-secret")
	
	if client.config.ClientID != "test-client-id" {
		t.Errorf("expected ClientID 'test-client-id', got: %s", client.config.ClientID)
	}
	
	if client.config.ClientSecret != "test-client-secret" {
		t.Errorf("expected ClientSecret 'test-client-secret', got: %s", client.config.ClientSecret)
	}
	
	if client.config.RedirectURL != RedirectURI {
		t.Errorf("expected RedirectURL '%s', got: %s", RedirectURI, client.config.RedirectURL)
	}
}

func TestGetAuthURL(t *testing.T) {
	client := NewSlackOAuthClient("test-client-id", "test-client-secret")
	
	authURL, state, err := client.GetAuthURL()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	
	if authURL == "" {
		t.Error("expected non-empty auth URL")
	}
	
	if state == "" {
		t.Error("expected non-empty state")
	}
	
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse auth URL: %v", err)
	}
	
	if parsedURL.Host != "slack.com" {
		t.Errorf("expected host 'slack.com', got: %s", parsedURL.Host)
	}
}

func TestGenerateRandomState(t *testing.T) {
	state1 := generateRandomState()
	state2 := generateRandomState()
	
	if state1 == state2 {
		t.Error("expected different states, but they were the same")
	}
	
	if len(state1) == 0 {
		t.Error("expected non-empty state")
	}
}

func TestCallbackServer(t *testing.T) {
	server := NewCallbackServer(0)
	
	t.Run("successful callback", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/callback?code=test-code&state=test-state", nil)
		w := httptest.NewRecorder()
		
		server.handleCallback(w, req)
		
		select {
		case result := <-server.result:
			if result.Code != "test-code" {
				t.Errorf("expected code 'test-code', got: %s", result.Code)
			}
			if result.State != "test-state" {
				t.Errorf("expected state 'test-state', got: %s", result.State)
			}
			if result.Error != "" {
				t.Errorf("expected no error, got: %s", result.Error)
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for callback result")
		}
		
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got: %d", w.Code)
		}
	})
	
	t.Run("error callback", func(t *testing.T) {
		server := NewCallbackServer(0)
		req := httptest.NewRequest("GET", "/callback?error=access_denied", nil)
		w := httptest.NewRecorder()
		
		server.handleCallback(w, req)
		
		select {
		case result := <-server.result:
			if result.Error != "access_denied" {
				t.Errorf("expected error 'access_denied', got: %s", result.Error)
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for callback result")
		}
		
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got: %d", w.Code)
		}
	})
}

func TestWaitForCallback(t *testing.T) {
	server := NewCallbackServer(0)
	
	t.Run("timeout", func(t *testing.T) {
		_, err := server.WaitForCallback(100 * time.Millisecond)
		if err == nil {
			t.Error("expected timeout error")
		}
	})
	
	t.Run("successful callback", func(t *testing.T) {
		server := NewCallbackServer(0)
		
		go func() {
			time.Sleep(50 * time.Millisecond)
			server.result <- CallbackResult{
				Code:  "test-code",
				State: "test-state",
			}
		}()
		
		result, err := server.WaitForCallback(time.Second)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		
		if result.Code != "test-code" {
			t.Errorf("expected code 'test-code', got: %s", result.Code)
		}
		if result.State != "test-state" {
			t.Errorf("expected state 'test-state', got: %s", result.State)
		}
	})
}