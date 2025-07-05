package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "slakctl",
	Short: "A CLI tool for managing Slack workspaces",
	Long:  "slakctl is a command-line tool for managing Slack workspaces using personal tokens.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(channelCmd)
	rootCmd.AddCommand(postCmd)
}