package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAuthCmd(t *testing.T) {
	t.Run("should require token argument or input", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:  "auth",
			RunE: runAuth,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("expected error when no token provided")
		}
	})

	t.Run("should accept token as argument", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "slakctl-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		cmd := &cobra.Command{
			Use:  "auth",
			RunE: runAuth,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"test-token"})

		err = cmd.Execute()
		if err != nil && !strings.Contains(err.Error(), "authentication failed") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestSearchCmd(t *testing.T) {
	t.Run("should require keyword argument", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:  "search",
			RunE: runSearch,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("expected error when no keyword provided")
		}
	})

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
			Use:  "search",
			RunE: runSearch,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"test"})

		err = cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "no authentication token found") {
			t.Errorf("expected authentication error, got: %v", err)
		}
	})
}

func TestPostCmd(t *testing.T) {
	t.Run("should require channel and message arguments", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:  "post",
			RunE: runPost,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		if err == nil {
			t.Error("expected error when no arguments provided")
		}
	})

	t.Run("should require both channel and message", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:  "post",
			RunE: runPost,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"#general"})

		err := cmd.Execute()
		if err == nil {
			t.Error("expected error when only channel provided")
		}
	})

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
			Use:  "post",
			RunE: runPost,
		}

		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"#general", "Hello, world!"})

		err = cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "no authentication token found") {
			t.Errorf("expected authentication error, got: %v", err)
		}
	})
}