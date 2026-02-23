package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetConfigPath tests the GetConfigPath function
func TestGetConfigPath(t *testing.T) {
	// Run multiple times to test idempotency
	for i := 0; i < 3; i++ {
		t.Run("iteration", func(t *testing.T) {
			path, err := GetConfigPath()
			if err != nil {
				t.Errorf("GetConfigPath() error = %v", err)
				return
			}

			// Check that path is not empty
			if path == "" {
				t.Error("GetConfigPath() returned empty string")
			}

			// Check that path contains expected components
			homeDir, err := os.UserHomeDir()
			if err != nil {
				t.Errorf("Failed to get home directory: %v", err)
				return
			}

			expectedDir := filepath.Join(homeDir, ".msa")
			expectedPath := filepath.Join(expectedDir, "config.json")

			if path != expectedPath {
				t.Errorf("GetConfigPath() = %v, want %v", path, expectedPath)
			}

			// Verify that the directory was created
			if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
				t.Errorf("GetConfigPath() did not create config directory: %s", expectedDir)
			}
		})
	}
}

// TestGetConfigPath_HomeDirError tests GetConfigPath when home directory cannot be determined
// Note: This test is difficult to implement without mocking os.UserHomeDir
// In practice, the function will succeed in most test environments
