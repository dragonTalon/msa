package siliconflow

import (
	"context"
	"msa/pkg/model"
	"testing"
)

// TestSiliconflowProvider_GetProvider tests the GetProvider method
func TestSiliconflowProvider_GetProvider(t *testing.T) {
	provider := SiliconflowProvider{}

	ctx := context.Background()
	got := provider.GetProvider(ctx)

	if got != model.Siliconflow {
		t.Errorf("GetProvider() = %v, want %v", got, model.Siliconflow)
	}
}

// TestSiliconflowProvider_Struct tests that SiliconflowProvider implements the interface
func TestSiliconflowProvider_Struct(t *testing.T) {
	var provider interface{} = SiliconflowProvider{}

	// Verify it has the GetProvider method
	if _, ok := provider.(interface {
		GetProvider(context.Context) model.LlmProvider
	}); !ok {
		t.Error("SiliconflowProvider should implement GetProvider method")
	}
}

// TestSiliconflowProvider_ListModels_Note documents that real API testing requires mocking
func TestSiliconflowProvider_ListModels_Note(t *testing.T) {
	t.Skip("Skipping real API test - requires HTTP mock setup and valid API credentials")

	// To properly test ListModels, we need to:
	// 1. Mock the HTTP client (utils.GetRestyClient())
	// 2. Set up test server responses for the API
	// 3. Provide test API credentials via config

	ctx := context.Background()
	provider := SiliconflowProvider{}

	// This would require mocking the HTTP client
	_, _ = provider.ListModels(ctx)
}

// TestModel_Struct tests the Model struct
func TestModel_Struct(t *testing.T) {
	model := Model{
		ID:      "test-model",
		Object:  "chat.completion",
		Created: 1234567890,
		OwnedBy: "siliconflow",
	}

	if model.ID != "test-model" {
		t.Errorf("Model ID = %v, want %v", model.ID, "test-model")
	}
	if model.Object != "chat.completion" {
		t.Errorf("Model Object = %v, want %v", model.Object, "chat.completion")
	}
}

// TestModelsResponse_Struct tests the ModelsResponse struct
func TestModelsResponse_Struct(t *testing.T) {
	response := ModelsResponse{
		Object: "list",
		Data: []Model{
			{ID: "model1", Object: "chat.completion"},
			{ID: "model2", Object: "chat.completion"},
		},
	}

	if response.Object != "list" {
		t.Errorf("ModelsResponse Object = %v, want %v", response.Object, "list")
	}
	if len(response.Data) != 2 {
		t.Errorf("ModelsResponse Data length = %v, want %v", len(response.Data), 2)
	}
}

// TestSILICONFLOW_BASE_URL tests the base URL constant
func TestSILICONFLOW_BASE_URL(t *testing.T) {
	if SILICONFLOW_BASE_URL == "" {
		t.Error("SILICONFLOW_BASE_URL should not be empty")
	}
	expectedURL := "https://api.siliconflow.cn/v1"
	if SILICONFLOW_BASE_URL != expectedURL {
		t.Errorf("SILICONFLOW_BASE_URL = %v, want %v", SILICONFLOW_BASE_URL, expectedURL)
	}
}
