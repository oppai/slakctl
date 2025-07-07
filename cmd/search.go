package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"slakctl/internal/config"
	"slakctl/internal/slack"

	"github.com/spf13/cobra"
)

var (
	searchCount    int
	searchFormat   string
	searchProgress bool
)

var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search for messages across channels",
	Long:  "Search for messages containing the specified keyword across all channels in the workspace.\n\nThis command supports pagination to fetch large numbers of results efficiently.",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().IntVarP(&searchCount, "count", "c", 20, "Number of messages to return (max 1000)")
	searchCmd.Flags().StringVarP(&searchFormat, "format", "f", "text", "Output format: text, json, or custom format string")
	searchCmd.Flags().BoolVarP(&searchProgress, "progress", "p", true, "Show progress during search")
}

func runSearch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("keyword argument is required")
	}

	// Validate count
	if searchCount < 1 || searchCount > 1000 {
		return fmt.Errorf("count must be between 1 and 1000")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		return fmt.Errorf("no authentication token found. Please run 'slakctl auth' first")
	}

	client := slack.NewClient(cfg.Token)

	keyword := args[0]

	options := slack.SearchOptions{
		MaxResults: searchCount,
	}

	var results *slack.SearchResult
	var err2 error

	if searchProgress {
		cmd.Println("Searching messages...")
		startTime := time.Now()

		options.ProgressFunc = func(current, total int) {
			elapsed := time.Since(startTime)
			if total > 0 {
				cmd.Printf("\rFound %d/%d messages (elapsed: %v)", current, total, elapsed.Round(time.Millisecond))
			} else {
				cmd.Printf("\rFound %d messages (elapsed: %v)", current, elapsed.Round(time.Millisecond))
			}
		}

		results, err2 = client.SearchWithOptions(keyword, options)

		if err2 == nil {
			cmd.Println() // 改行
		}
	} else {
		results, err2 = client.SearchWithOptions(keyword, options)
	}

	if err2 != nil {
		return fmt.Errorf("search failed: %w", err2)
	}

	if len(results.Matches) == 0 {
		if searchFormat == "json" {
			emptyResult := map[string]interface{}{
				"matches": []interface{}{},
				"total":   0,
				"query":   keyword,
			}
			jsonOutput, _ := json.MarshalIndent(emptyResult, "", "  ")
			cmd.Println(string(jsonOutput))
		} else {
			cmd.Printf("No messages found containing '%s'\n", keyword)
		}
		return nil
	}

	return formatSearchResults(cmd, results, keyword, searchFormat)
}

func formatSearchResults(cmd *cobra.Command, results *slack.SearchResult, keyword, format string) error {
	switch format {
	case "json":
		output := map[string]interface{}{
			"matches": results.Matches,
			"total":   results.Total,
			"query":   keyword,
		}
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(jsonOutput))

	case "text":
		cmd.Printf("Found %d messages containing '%s' (total: %d):\n\n", len(results.Matches), keyword, results.Total)

		for _, msg := range results.Matches {
			channelName := msg.Channel.Name
			if channelName == "" {
				channelName = msg.Channel.ID
			}

			username := msg.Username
			if username == "" {
				username = msg.User
			}

			cmd.Printf("Channel: #%s\n", channelName)
			cmd.Printf("User: %s\n", username)
			cmd.Printf("Text: %s\n", strings.TrimSpace(msg.Text))
			cmd.Printf("Timestamp: %s\n", msg.TS)
			if msg.Permalink != "" {
				cmd.Printf("Link: %s\n", msg.Permalink)
			}
			cmd.Println("---")
		}

	default:
		// Custom format string - jq-like formatting
		for _, msg := range results.Matches {
			output := formatMessage(msg, format)
			cmd.Println(output)
		}
	}

	return nil
}

func formatMessage(msg slack.Message, format string) string {
	channelName := msg.Channel.Name
	if channelName == "" {
		channelName = msg.Channel.ID
	}

	username := msg.Username
	if username == "" {
		username = msg.User
	}

	// Simple template replacement
	output := format
	output = strings.ReplaceAll(output, "{channel}", channelName)
	output = strings.ReplaceAll(output, "{user}", username)
	output = strings.ReplaceAll(output, "{text}", strings.TrimSpace(msg.Text))
	output = strings.ReplaceAll(output, "{timestamp}", msg.TS)
	output = strings.ReplaceAll(output, "{permalink}", msg.Permalink)
	output = strings.ReplaceAll(output, "{channel_id}", msg.Channel.ID)
	output = strings.ReplaceAll(output, "{user_id}", msg.User)
	output = strings.ReplaceAll(output, "\\n", "\n")
	output = strings.ReplaceAll(output, "\\t", "\t")

	return output
}
