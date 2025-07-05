package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

func (c *Client) makeRequest(method, endpoint string, data interface{}) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("https://slack.com/api/%s", endpoint), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func (c *Client) TestAuth() error {
	body, err := c.makeRequest("GET", "auth.test", nil)
	if err != nil {
		return err
	}

	var response struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.OK {
		return fmt.Errorf("authentication failed: %s", response.Error)
	}

	return nil
}

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) ListChannels() ([]Channel, error) {
	body, err := c.makeRequest("GET", "conversations.list", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		OK       bool      `json:"ok"`
		Error    string    `json:"error,omitempty"`
		Channels []Channel `json:"channels"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.OK {
		return nil, fmt.Errorf("failed to list channels: %s", response.Error)
	}

	return response.Channels, nil
}

type SearchResult struct {
	Matches    []Message `json:"matches"`
	Total      int       `json:"total"`
	Pagination struct {
		TotalCount int `json:"total_count"`
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		PageCount  int `json:"page_count"`
		First      int `json:"first"`
		Last       int `json:"last"`
	} `json:"paging"`
}

type Message struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	User      string `json:"user"`
	Username  string `json:"username"`
	Channel   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	TS        string `json:"ts"`
	Permalink string `json:"permalink"`
}

func (c *Client) Search(query string) (*SearchResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("sort", "timestamp")
	params.Set("sort_dir", "desc")
	params.Set("count", "20")

	endpoint := "search.messages?" + params.Encode()
	body, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		OK       bool         `json:"ok"`
		Error    string       `json:"error,omitempty"`
		Messages SearchResult `json:"messages"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.OK {
		return nil, fmt.Errorf("search failed: %s", response.Error)
	}

	return &response.Messages, nil
}

func (c *Client) PostMessage(channel, text string) error {
	channelID := channel
	if strings.HasPrefix(channel, "#") {
		channelID = strings.TrimPrefix(channel, "#")
	}

	data := map[string]interface{}{
		"channel": channelID,
		"text":    text,
	}

	body, err := c.makeRequest("POST", "chat.postMessage", data)
	if err != nil {
		return err
	}

	var response struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !response.OK {
		return fmt.Errorf("failed to post message: %s", response.Error)
	}

	return nil
}