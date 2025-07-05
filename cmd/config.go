package cmd

import (
	"fmt"

	"slakctl/internal/auth"
	"slakctl/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Slack app credentials",
	Long:  "Configure Slack app credentials for OAuth2 authentication.",
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set Slack app credentials",
	Long:  "Set Slack app Client ID and Client Secret for OAuth2 authentication.",
	RunE:  runConfigSet,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Show current Slack app configuration (Client ID only, for security).",
	RunE:  runConfigShow,
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	var clientID, clientSecret string

	fmt.Print("Enter Slack App Client ID: ")
	if _, err := fmt.Scanln(&clientID); err != nil {
		return fmt.Errorf("failed to read client ID: %w", err)
	}

	fmt.Print("Enter Slack App Client Secret: ")
	if _, err := fmt.Scanln(&clientSecret); err != nil {
		return fmt.Errorf("failed to read client secret: %w", err)
	}

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("client ID and client secret cannot be empty")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cfg.ClientID = clientID
	cfg.ClientSecret = clientSecret

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Slack app credentials saved successfully!")
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cmd.Printf("Client ID: %s\n", cfg.ClientID)
	if cfg.ClientSecret != "" {
		cmd.Printf("Client Secret: %s\n", maskSecret(cfg.ClientSecret))
	} else {
		cmd.Println("Client Secret: Not set")
	}
	
	if cfg.Token != "" {
		cmd.Printf("Token: %s\n", maskSecret(cfg.Token))
	} else {
		cmd.Println("Token: Not set")
	}

	return nil
}

func maskSecret(secret string) string {
	if len(secret) < 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

var configOAuthCmd = &cobra.Command{
	Use:   "oauth",
	Short: "Authenticate using OAuth2 flow",
	Long:  "Authenticate with Slack using OAuth2 flow. This will open a browser for authentication.",
	RunE:  runAuthOAuth,
}

func runAuthOAuth(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return fmt.Errorf("Slack app credentials not configured. Please run 'slakctl config set' first")
	}

	token, err := auth.StartOAuthFlow(cfg.ClientID, cfg.ClientSecret)
	if err != nil {
		return fmt.Errorf("OAuth authentication failed: %w", err)
	}

	cfg.Token = token

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("OAuth authentication successful! Token saved.")
	return nil
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configOAuthCmd)
}