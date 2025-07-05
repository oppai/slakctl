package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestChannelListCmd(t *testing.T) {
	t.Run("should require authentication", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "slakctl-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		cmd := &cobra.Command{
			Use:  "list",
			RunE: runChannelList,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err = cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "no authentication token found") {
			t.Errorf("expected authentication error, got: %v", err)
		}
	})
}