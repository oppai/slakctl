package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestConfigSetCmd(t *testing.T) {
	t.Run("should require client credentials", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "slakctl-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		cmd := &cobra.Command{
			Use:  "set",
			RunE: runConfigSet,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err = cmd.Execute()
		if err == nil {
			t.Error("expected error when no input provided")
		}
	})
}

func TestConfigShowCmd(t *testing.T) {
	t.Run("should show empty config", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "slakctl-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		cmd := &cobra.Command{
			Use:  "show",
			RunE: runConfigShow,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err = cmd.Execute()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Client ID:") {
			t.Errorf("expected output to contain 'Client ID:', got: %s", output)
		}
		if !strings.Contains(output, "Client Secret:") {
			t.Errorf("expected output to contain 'Client Secret:', got: %s", output)
		}
		if !strings.Contains(output, "Token:") {
			t.Errorf("expected output to contain 'Token:', got: %s", output)
		}
	})
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "****"},
		{"short", "****"},
		{"12345678", "1234****5678"},
		{"xoxb-test-123456789-987654321-wxyz", "xoxb****wxyz"},
	}

	for _, test := range tests {
		result := maskSecret(test.input)
		if result != test.expected {
			t.Errorf("maskSecret(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestAuthOAuthCmd(t *testing.T) {
	t.Run("should require client credentials", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "slakctl-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		cmd := &cobra.Command{
			Use:  "oauth",
			RunE: runAuthOAuth,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err = cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "Slack app credentials not configured") {
			t.Errorf("expected credentials error, got: %v", err)
		}
	})
}
