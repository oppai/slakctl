package cmd

import (
	"fmt"

	"slakctl/internal/config"
	"slakctl/internal/slack"

	"github.com/spf13/cobra"
)

var postCmd = &cobra.Command{
	Use:   "post [channel] [message]",
	Short: "Post a message to a channel",
	Long:  "Post a message to the specified channel. Channel can be specified with or without the # prefix.",
	Args:  cobra.ExactArgs(2),
	RunE:  runPost,
}

func runPost(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("both channel and message arguments are required")
	}
	
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		return fmt.Errorf("no authentication token found. Please run 'slakctl auth' first")
	}

	client := slack.NewClient(cfg.Token)
	
	channel := args[0]
	message := args[1]

	if err := client.PostMessage(channel, message); err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	fmt.Printf("Message posted successfully to %s\n", channel)
	return nil
}