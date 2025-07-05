package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

const (
	SlackAuthURL  = "https://slack.com/oauth/v2/authorize"
	SlackTokenURL = "https://slack.com/api/oauth.v2.access"
	RedirectURI   = "https://b1c3-42-148-67-84.ngrok-free.app/callback"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

type SlackOAuthClient struct {
	config *oauth2.Config
}

func NewSlackOAuthClient(clientID, clientSecret string) *SlackOAuthClient {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  RedirectURI,
		Scopes:       []string{"channels:history", "channels:read", "channels:write", "chat:write", "search:read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  SlackAuthURL,
			TokenURL: SlackTokenURL,
		},
	}

	return &SlackOAuthClient{config: config}
}

func (c *SlackOAuthClient) GetAuthURL() (string, string, error) {
	state := generateRandomState()
	authURL := c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL, state, nil
}

func (c *SlackOAuthClient) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.config.Exchange(ctx, code)
}

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

type CallbackServer struct {
	server *http.Server
	result chan CallbackResult
}

type CallbackResult struct {
	Code  string
	State string
	Error string
}

func NewCallbackServer(port int) *CallbackServer {
	return &CallbackServer{
		result: make(chan CallbackResult, 1),
	}
}

func (s *CallbackServer) Start(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request to unexpected path: %s %s\n", r.Method, r.URL.String())
		if r.URL.Path == "/callback" {
			s.handleCallback(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not found: %s", r.URL.Path)
		}
	})
	
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	fmt.Printf("Starting callback server on port %d...\n", port)
	return s.server.ListenAndServe()
}

func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Callback received: %s %s\n", r.Method, r.URL.String())
	fmt.Printf("Raw query: %s\n", r.URL.RawQuery)
	
	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	errorParam := query.Get("error")

	fmt.Printf("Parsed parameters - Code: '%s', State: '%s', Error: '%s'\n", code, state, errorParam)

	result := CallbackResult{
		Code:  code,
		State: state,
		Error: errorParam,
	}

	select {
	case s.result <- result:
	default:
	}

	if errorParam != "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Error</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .error { color: red; }
    </style>
</head>
<body>
    <h1>Authentication Error</h1>
    <p class="error">Error: %s</p>
    <p>Please try again.</p>
</body>
</html>`, errorParam)
	} else {
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Success</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .success { color: green; }
    </style>
</head>
<body>
    <h1>Authentication Successful!</h1>
    <p class="success">You can now close this window and return to the terminal.</p>
</body>
</html>`)
	}
}

func (s *CallbackServer) WaitForCallback(timeout time.Duration) (CallbackResult, error) {
	select {
	case result := <-s.result:
		return result, nil
	case <-time.After(timeout):
		return CallbackResult{}, fmt.Errorf("timeout waiting for callback")
	}
}

func (s *CallbackServer) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}

func StartOAuthFlow(clientID, clientSecret string) (string, error) {
	client := NewSlackOAuthClient(clientID, clientSecret)
	
	authURL, state, err := client.GetAuthURL()
	if err != nil {
		return "", fmt.Errorf("failed to get auth URL: %w", err)
	}

	fmt.Printf("Opening browser to: %s\n", authURL)
	if err := OpenBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser automatically. Please manually open: %s\n", authURL)
	}

	server := NewCallbackServer(8090)
	go func() {
		if err := server.Start(8090); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start callback server: %v\n", err)
		}
	}()

	defer server.Stop()

	fmt.Println("Waiting for authentication callback...")
	fmt.Printf("Make sure ngrok is running: ngrok http 8090\n")
	fmt.Printf("Expected ngrok URL should be configured in Slack app: %s\n", RedirectURI)
	
	result, err := server.WaitForCallback(5 * time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to receive callback: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("authentication error: %s", result.Error)
	}

	fmt.Printf("Received callback - Code: %s, State: %s\n", result.Code, result.State)
	
	if result.State != state {
		return "", fmt.Errorf("state mismatch: expected %s, got %s", state, result.State)
	}

	token, err := client.ExchangeCodeForToken(context.Background(), result.Code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token.AccessToken, nil
}