package cmd

import (
	"fmt"

	"slakctl/internal/config"
	"slakctl/internal/slack"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Slack",
	Long:  "Authenticate with Slack using either a personal token or OAuth2 flow.",
}

var authTokenCmd = &cobra.Command{
	Use:   "token [token]",
	Short: "Authenticate with Slack using a personal token",
	Long:  "Set up authentication with Slack using a personal token. If no token is provided, you will be prompted to enter one.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runAuth,
}

func runAuth(cmd *cobra.Command, args []string) error {
	var token string

	if len(args) > 0 {
		token = args[0]
	} else {
		fmt.Print("Enter your Slack token: ")
		if _, err := fmt.Scanln(&token); err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
	}

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	client := slack.NewClient(token)
	if err := client.TestAuth(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	cfg := &config.Config{
		Token: token,
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Authentication successful! Token saved.")
	return nil
}

func init() {
	authCmd.AddCommand(authTokenCmd)
	authCmd.Flags().BoolP("help", "h", false, "Help for auth command")
}