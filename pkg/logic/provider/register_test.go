package provider

import (
	"context"
	"msa/pkg/model"
	"testing"
)

// mockProvider is a mock implementation of LLMProvider for testing
type mockProvider struct {
	providerType model.LlmProvider
}

func (m *mockProvider) GetProvider(ctx context.Context) model.LlmProvider {
	return m.providerType
}

func (m *mockProvider) ListModels(ctx context.Context) ([]*model.LLMModel, error) {
	return []*model.LLMModel{
		{Name: "model1", Description: "Test model 1"},
		{Name: "model2", Description: "Test model 2"},
	}, nil
}

// helper function to create a mock provider
func newMockProvider(providerType model.LlmProvider) *mockProvider {
	return &mockProvider{providerType: providerType}
}

// TestRegisterProvider tests the RegisterProvider function
func TestRegisterProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider model.LlmProvider
		instance LLMProvider
	}{
		{
			name:     "register mock provider",
			provider: "mock_provider",
			instance: newMockProvider("mock_provider"),
		},
		{
			name:     "register test provider",
			provider: "test_provider",
			instance: newMockProvider("test_provider"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Register the provider
			RegisterProvider(tt.provider, tt.instance)

			// Verify provider is in providerMap
			got := providerMap[tt.provider]
			if got == nil {
				t.Errorf("RegisterProvider() provider not found in map")
			}
			if got != nil && got.GetProvider(context.Background()) != tt.provider {
				t.Errorf("RegisterProvider() provider type = %v, want %v", got.GetProvider(context.Background()), tt.provider)
			}
		})
	}
}

// TestLLMProvider_Interface tests that mockProvider correctly implements LLMProvider
func TestLLMProvider_Interface(t *testing.T) {
	ctx := context.Background()
	provider := newMockProvider(model.Siliconflow)

	// Test GetProvider
	got := provider.GetProvider(ctx)
	if got != model.Siliconflow {
		t.Errorf("GetProvider() = %v, want %v", got, model.Siliconflow)
	}

	// Test ListModels
	models, err := provider.ListModels(ctx)
	if err != nil {
		t.Errorf("ListModels() error = %v", err)
	}
	if len(models) != 2 {
		t.Errorf("ListModels() returned %d models, want 2", len(models))
	}
	if models[0].Name != "model1" {
		t.Errorf("ListModels()[0].Name = %v, want model1", models[0].Name)
	}
}

// TestGetProvider_WithConfig tests GetProvider with different config scenarios
func TestGetProvider_WithConfig(t *testing.T) {
	// Note: This test is limited because GetProvider depends on
	// config.GetLocalStoreConfig() which uses global state and file system
	// A full integration test would require setting up test config files

	// For now, we just verify the function doesn't panic
	provider := GetProvider()
	_ = provider // provider may be nil if config is not set
}

// TestGetProvider_NotFound tests GetProvider when provider is not registered
func TestGetProvider_NotFound(t *testing.T) {
	// This test assumes there's a way to set an unregistered provider
	// Since we can't easily modify the config in tests,
	// we just document that GetProvider returns nil for unregistered providers
}

// TestProviderMap_GlobalState tests that providerMap is a package-level variable
func TestProviderMap_GlobalState(t *testing.T) {
	// Verify that providerMap is accessible (tests the package structure)
	if providerMap == nil {
		t.Error("providerMap should be initialized")
	}

	// Register a test provider
	testProvider := newMockProvider("global_state_test")
	RegisterProvider("global_state_test", testProvider)

	// Verify it's accessible
	got := providerMap["global_state_test"]
	if got == nil {
		t.Error("Provider should be in providerMap after registration")
	}
}

// TestSiliconflowProvider_Registered tests that SiliconflowProvider is registered
func TestSiliconflowProvider_Registered(t *testing.T) {
	// The SiliconflowProvider should be registered via init()
	// This test verifies it's available in the providerMap
	provider := providerMap[model.Siliconflow]
	if provider == nil {
		t.Error("SiliconflowProvider should be registered via init()")
	}

	// If registered, verify it implements the interface
	if provider != nil {
		providerType := provider.GetProvider(context.Background())
		if providerType != model.Siliconflow {
			t.Errorf("Registered provider type = %v, want %v", providerType, model.Siliconflow)
		}
	}
}
