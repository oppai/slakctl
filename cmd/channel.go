package cmd

import (
	"fmt"
	"time"

	"slakctl/internal/config"
	"slakctl/internal/slack"

	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Channel management commands",
	Long:  "Commands for managing and listing channels in your Slack workspace.",
}

var (
	showProgress    bool
	allChannels     bool
	includeArchived bool
)

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List channels",
	Long:  "List channels in the workspace that the authenticated user has access to.\n\nBy default, this command fetches up to 1000 channels and excludes archived channels. Use --all to fetch all channels and --archived to include archived channels.",
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

	var channels []slack.Channel
	var err2 error

	options := slack.ListChannelsOptions{
		AllChannels:     allChannels,
		IncludeArchived: includeArchived,
	}

	if showProgress {
		fmt.Println("Fetching channels...")
		startTime := time.Now()

		options.ProgressFunc = func(current, total int) {
			elapsed := time.Since(startTime)
			if total > 0 {
				fmt.Printf("\rFetched %d/%d channels (elapsed: %v)", current, total, elapsed.Round(time.Millisecond))
			} else {
				fmt.Printf("\rFetched %d channels (elapsed: %v)", current, elapsed.Round(time.Millisecond))
			}
		}

		channels, err2 = client.ListChannelsWithOptions(options)

		if err2 == nil {
			fmt.Println() // 改行
		}
	} else {
		channels, err2 = client.ListChannelsWithOptions(options)
	}

	if err2 != nil {
		return fmt.Errorf("failed to list channels: %w", err2)
	}

	if len(channels) == 0 {
		fmt.Println("No channels found")
		return nil
	}

	fmt.Printf("Found %d channels:\n\n", len(channels))

	for _, channel := range channels {
		fmt.Printf("ID: %s\n", channel.ID)
		fmt.Printf("Name: #%s\n", channel.Name)
		if channel.IsArchived {
			fmt.Printf("Status: Archived\n")
		}
		fmt.Println("---")
	}

	return nil
}

func init() {
	channelListCmd.Flags().BoolVarP(&showProgress, "progress", "p", true, "Show progress during channel listing")
	channelListCmd.Flags().BoolVarP(&allChannels, "all", "a", false, "Fetch all channels (default: limit to 1000)")
	channelListCmd.Flags().BoolVarP(&includeArchived, "archived", "", false, "Include archived channels")
	channelCmd.AddCommand(channelListCmd)
}
