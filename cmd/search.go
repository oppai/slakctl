package cmd

import (
	"fmt"
	"strings"

	"slakctl/internal/config"
	"slakctl/internal/slack"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search for messages across channels",
	Long:  "Search for messages containing the specified keyword across all channels in the workspace.",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("keyword argument is required")
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
	fmt.Printf("Searching for '%s'...\n", keyword)
	
	results, err := client.Search(keyword)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results.Matches) == 0 {
		fmt.Printf("No messages found containing '%s'\n", keyword)
		return nil
	}

	fmt.Printf("Found %d messages containing '%s' (total: %d):\n\n", len(results.Matches), keyword, results.Total)
	
	for _, msg := range results.Matches {
		channelName := msg.Channel.Name
		if channelName == "" {
			channelName = msg.Channel.ID
		}
		
		username := msg.Username
		if username == "" {
			username = msg.User
		}
		
		fmt.Printf("Channel: #%s\n", channelName)
		fmt.Printf("User: %s\n", username)
		fmt.Printf("Text: %s\n", strings.TrimSpace(msg.Text))
		fmt.Printf("Timestamp: %s\n", msg.TS)
		if msg.Permalink != "" {
			fmt.Printf("Link: %s\n", msg.Permalink)
		}
		fmt.Println("---")
	}

	return nil
}