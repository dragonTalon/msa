package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"msa/pkg/model"
)

// helper function to create a temporary directory for test config
func setupTestConfigDir(t *testing.T) (testDir string, cleanup func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "msa-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Override HOME environment variable for this test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	cleanup = func() {
		os.Setenv("HOME", oldHome)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// TestLocalStoreConfig_Struct tests the LocalStoreConfig structure
func TestLocalStoreConfig_Struct(t *testing.T) {
	testConfig := &LocalStoreConfig{
		APIKey:   "test-key",
		BaseURL:  "http://test.com",
		Provider: model.Siliconflow,
		Model:    "gpt-4",
		LogConfig: &LogConfig{
			Level:      "debug",
			Format:     "json",
			Output:     "stdout",
			ShowColor:  true,
			TimeFormat: "2006-01-02 15:04:05",
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Test JSON deserialization
	var decoded LocalStoreConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if decoded.APIKey != testConfig.APIKey {
		t.Errorf("APIKey = %v, want %v", decoded.APIKey, testConfig.APIKey)
	}
	if decoded.BaseURL != testConfig.BaseURL {
		t.Errorf("BaseURL = %v, want %v", decoded.BaseURL, testConfig.BaseURL)
	}
	if decoded.Provider != testConfig.Provider {
		t.Errorf("Provider = %v, want %v", decoded.Provider, testConfig.Provider)
	}
	if decoded.Model != testConfig.Model {
		t.Errorf("Model = %v, want %v", decoded.Model, testConfig.Model)
	}
}

// TestInitConfig_CreateDefault tests InitConfig creates default config when none exists
func TestInitConfig_CreateDefault(t *testing.T) {
	testDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Ensure config directory doesn't exist
	configDir := filepath.Join(testDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Run InitConfig - should create default config file
	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(configDir, CONFIG_FILE)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("InitConfig() did not create config file")
	}

	// Read and verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var cfg LocalStoreConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify default values
	if cfg.Provider != model.Siliconflow {
		t.Errorf("Default Provider = %v, want %v", cfg.Provider, model.Siliconflow)
	}
	if cfg.LogConfig == nil {
		t.Error("Default LogConfig should not be nil")
	} else {
		if cfg.LogConfig.Level != "info" {
			t.Errorf("Default LogConfig.Level = %v, want 'info'", cfg.LogConfig.Level)
		}
		if cfg.LogConfig.Format != "text" {
			t.Errorf("Default LogConfig.Format = %v, want 'text'", cfg.LogConfig.Format)
		}
	}
}

// TestReloadConfig_WithValidConfig tests ReloadConfig with a valid config file
func TestReloadConfig_WithValidConfig(t *testing.T) {
	testDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Create config directory
	configDir := filepath.Join(testDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create a valid config file
	testConfig := &LocalStoreConfig{
		APIKey:   "test-api-key",
		BaseURL:  "http://test.example.com",
		Provider: model.Siliconflow,
		Model:    "gpt-4",
		LogConfig: &LogConfig{
			Level:      "debug",
			Format:     "json",
			Output:     "file",
			File:       "/tmp/test.log",
			ShowColor:  false,
			TimeFormat: "2006-01-02 15:04:05",
		},
	}

	configPath := filepath.Join(configDir, CONFIG_FILE)
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Reload config
	cfg := ReloadConfig()

	// Verify loaded values
	if cfg.APIKey != testConfig.APIKey {
		t.Errorf("ReloadConfig() APIKey = %v, want %v", cfg.APIKey, testConfig.APIKey)
	}
	if cfg.BaseURL != testConfig.BaseURL {
		t.Errorf("ReloadConfig() BaseURL = %v, want %v", cfg.BaseURL, testConfig.BaseURL)
	}
	if cfg.Provider != testConfig.Provider {
		t.Errorf("ReloadConfig() Provider = %v, want %v", cfg.Provider, testConfig.Provider)
	}
	if cfg.Model != testConfig.Model {
		t.Errorf("ReloadConfig() Model = %v, want %v", cfg.Model, testConfig.Model)
	}
	if cfg.LogConfig == nil {
		t.Error("ReloadConfig() LogConfig should not be nil")
	}
}

// TestReloadConfig_MissingFile tests ReloadConfig when config file doesn't exist
func TestReloadConfig_MissingFile(t *testing.T) {
	testDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Create config directory but not the file
	configDir := filepath.Join(testDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Reload should return empty config, not error
	cfg := ReloadConfig()
	if cfg == nil {
		t.Error("ReloadConfig() should return empty config, not nil")
	}
	if cfg.APIKey != "" {
		t.Errorf("ReloadConfig() with missing file should have empty APIKey, got %v", cfg.APIKey)
	}
}

// TestReloadConfig_InvalidJSON tests ReloadConfig with invalid JSON
func TestReloadConfig_InvalidJSON(t *testing.T) {
	testDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Create config directory
	configDir := filepath.Join(testDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create invalid JSON file
	configPath := filepath.Join(configDir, CONFIG_FILE)
	invalidJSON := []byte("{invalid json content")
	if err := os.WriteFile(configPath, invalidJSON, 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Reload should return empty config, not panic
	cfg := ReloadConfig()
	if cfg == nil {
		t.Error("ReloadConfig() should return empty config, not nil")
	}
}

// TestSaveConfig tests the SaveConfig function
func TestSaveConfig(t *testing.T) {
	testDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Create config directory
	configDir := filepath.Join(testDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Set up a config to save
	testConfig := &LocalStoreConfig{
		APIKey:   "saved-api-key",
		BaseURL:  "http://saved.example.com",
		Provider: model.Siliconflow,
		Model:    "claude-3",
	}

	// Note: SaveConfig uses configCache, so we need to set it
	// For this test, we'll verify the function doesn't error
	// In a real scenario, you'd need to reset the package state

	// Create the config file first to ensure SaveConfig has something to write
	configPath := filepath.Join(configDir, CONFIG_FILE)
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Reload to populate cache
	_ = ReloadConfig()

	// Save the config
	if err := SaveConfig(); err != nil {
		t.Errorf("SaveConfig() error = %v", err)
	}

	// Verify file still exists and is valid
	savedData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	var savedCfg LocalStoreConfig
	if err := json.Unmarshal(savedData, &savedCfg); err != nil {
		t.Fatalf("Failed to unmarshal saved config: %v", err)
	}

	if savedCfg.APIKey != testConfig.APIKey {
		t.Errorf("Saved APIKey = %v, want %v", savedCfg.APIKey, testConfig.APIKey)
	}
}

// TestGetLocalStoreConfig_Caching tests that GetLocalStoreConfig caches results
func TestGetLocalStoreConfig_Caching(t *testing.T) {
	testDir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	// Create config directory and file
	configDir := filepath.Join(testDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	testConfig := &LocalStoreConfig{
		APIKey:  "cache-test",
		BaseURL: "http://cache.test",
	}

	configPath := filepath.Join(configDir, CONFIG_FILE)
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// First call
	cfg1 := GetLocalStoreConfig()
	if cfg1 == nil {
		t.Fatal("GetLocalStoreConfig() returned nil")
	}

	// Modify the file
	testConfig.APIKey = "modified"
	data, _ = json.MarshalIndent(testConfig, "", "  ")
	os.WriteFile(configPath, data, 0644)

	// Second call - due to sync.Once, should return cached value
	cfg2 := GetLocalStoreConfig()
	if cfg2 == nil {
		t.Fatal("GetLocalStoreConfig() returned nil on second call")
	}

	// Cached values should be the same (not the modified value)
	if cfg1.APIKey != cfg2.APIKey {
		t.Errorf("GetLocalStoreConfig() caching failed: first call got %v, second got %v", cfg1.APIKey, cfg2.APIKey)
	}
}

// TestLogConfig_DefaultConfig tests the DefaultConfig function
func TestLogConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	tests := []struct {
		field string
		want  string
		got   string
	}{
		{"Level", "info", cfg.Level},
		{"Format", "text", cfg.Format},
		{"Output", "file", cfg.Output},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("DefaultConfig() %s = %v, want %v", tt.field, tt.got, tt.want)
		}
	}

	if !cfg.ShowColor {
		t.Error("DefaultConfig() ShowColor should be true by default")
	}
	if cfg.TimeFormat == "" {
		t.Error("DefaultConfig() TimeFormat should not be empty")
	}
}
