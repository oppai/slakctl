package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsArchived bool   `json:"is_archived"`
}

type ResponseMetadata struct {
	NextCursor string `json:"next_cursor"`
}

type ListChannelsOptions struct {
	AllChannels     bool
	IncludeArchived bool
	ProgressFunc    func(current, total int)
}

type SearchOptions struct {
	MaxResults   int
	ProgressFunc func(current, total int)
}

func (c *Client) ListChannels() ([]Channel, error) {
	return c.ListChannelsWithOptions(ListChannelsOptions{})
}

func (c *Client) ListChannelsWithProgress(progressFunc func(current, total int)) ([]Channel, error) {
	return c.ListChannelsWithOptions(ListChannelsOptions{
		AllChannels:  true,
		ProgressFunc: progressFunc,
	})
}

func (c *Client) ListChannelsWithOptions(options ListChannelsOptions) ([]Channel, error) {
	var allChannels []Channel
	cursor := ""
	pageCount := 0
	maxChannels := 1000

	if options.AllChannels {
		maxChannels = 0 // 0は無制限を意味する
	}

	for {
		pageCount++
		params := url.Values{}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		// 最大1000件ずつ取得
		params.Set("limit", "1000")

		// アーカイブされたチャンネルを除外（APIレベルでフィルタリング）
		if !options.IncludeArchived {
			params.Set("exclude_archived", "true")
		}

		endpoint := "conversations.list"
		if len(params) > 0 {
			endpoint += "?" + params.Encode()
		}

		body, err := c.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		var response struct {
			OK               bool             `json:"ok"`
			Error            string           `json:"error,omitempty"`
			Channels         []Channel        `json:"channels"`
			ResponseMetadata ResponseMetadata `json:"response_metadata"`
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if !response.OK {
			return nil, fmt.Errorf("failed to list channels: %s", response.Error)
		}

		allChannels = append(allChannels, response.Channels...)

		// プログレス表示
		if options.ProgressFunc != nil {
			options.ProgressFunc(len(allChannels), maxChannels)
		}

		// 1000件制限に達した場合
		if maxChannels > 0 && len(allChannels) >= maxChannels {
			allChannels = allChannels[:maxChannels]
			break
		}

		// 次のページがあるかチェック
		if response.ResponseMetadata.NextCursor == "" {
			break
		}

		// 1000件取得ごとに3秒のインターバル（最初のページ以外）
		if pageCount > 1 && options.AllChannels {
			time.Sleep(3 * time.Second)
		}

		cursor = response.ResponseMetadata.NextCursor
	}

	return allChannels, nil
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
	Type     string `json:"type"`
	Text     string `json:"text"`
	User     string `json:"user"`
	Username string `json:"username"`
	Channel  struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	TS        string `json:"ts"`
	Permalink string `json:"permalink"`
}

func (c *Client) Search(query string) (*SearchResult, error) {
	return c.SearchWithOptions(query, SearchOptions{MaxResults: 20})
}

func (c *Client) SearchWithCount(query string, count int) (*SearchResult, error) {
	return c.SearchWithOptions(query, SearchOptions{MaxResults: count})
}

func (c *Client) SearchWithOptions(query string, options SearchOptions) (*SearchResult, error) {
	var allMatches []Message
	page := 1
	maxResults := options.MaxResults
	if maxResults <= 0 {
		maxResults = 20 // デフォルト
	}

	for {
		params := url.Values{}
		params.Set("query", query)
		params.Set("sort", "timestamp")
		params.Set("sort_dir", "desc")
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("count", "100") // 1ページあたり最大100件

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

		// 新しいマッチを追加
		newMatches := response.Messages.Matches
		remainingSlots := maxResults - len(allMatches)

		if remainingSlots <= 0 {
			break
		}

		if len(newMatches) > remainingSlots {
			newMatches = newMatches[:remainingSlots]
		}

		allMatches = append(allMatches, newMatches...)

		// プログレス表示
		if options.ProgressFunc != nil {
			options.ProgressFunc(len(allMatches), maxResults)
		}

		// 制限に達した場合
		if len(allMatches) >= maxResults {
			break
		}

		// 次のページがあるかチェック
		if page >= response.Messages.Pagination.PageCount {
			break
		}

		// ページ間のインターバル（最初のページ以外）
		if page > 1 {
			time.Sleep(1 * time.Second)
		}

		page++
	}

	// 結果を構築
	result := &SearchResult{
		Matches: allMatches,
		Total:   len(allMatches),
		Pagination: struct {
			TotalCount int `json:"total_count"`
			Page       int `json:"page"`
			PerPage    int `json:"per_page"`
			PageCount  int `json:"page_count"`
			First      int `json:"first"`
			Last       int `json:"last"`
		}{
			TotalCount: len(allMatches),
			Page:       1,
			PerPage:    len(allMatches),
			PageCount:  1,
			First:      1,
			Last:       len(allMatches),
		},
	}

	return result, nil
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
