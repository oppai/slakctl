package cmd

import (
	"fmt"

	"slakctl/internal/config"
	"slakctl/internal/slack"

	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Channel management commands",
	Long:  "Commands for managing and listing channels in your Slack workspace.",
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all channels",
	Long:  "List all channels in the workspace that the authenticated user has access to.",
	RunE:  runChannelList,
}

func runChannelList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		return fmt.Errorf("no authentication token found. Please run 'slakctl auth' first")
	}

	client := slack.NewClient(cfg.Token)
	
	channels, err := client.ListChannels()
	if err != nil {
		return fmt.Errorf("failed to list channels: %w", err)
	}

	if len(channels) == 0 {
		fmt.Println("No channels found")
		return nil
	}

	fmt.Printf("Found %d channels:\n\n", len(channels))
	
	for _, channel := range channels {
		fmt.Printf("ID: %s\n", channel.ID)
		fmt.Printf("Name: #%s\n", channel.Name)
		fmt.Println("---")
	}

	return nil
}

func init() {
	channelCmd.AddCommand(channelListCmd)
}