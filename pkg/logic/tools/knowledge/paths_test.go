package knowledge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetKnowledgeDir(t *testing.T) {
	path, err := GetKnowledgeDir()
	if err != nil {
		t.Fatalf("GetKnowledgeDir() error = %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, KnowledgeDir)
	if path != expected {
		t.Errorf("GetKnowledgeDir() = %v, want %v", path, expected)
	}
}

func TestGetErrorsPath(t *testing.T) {
	path, err := GetErrorsPath()
	if err != nil {
		t.Fatalf("GetErrorsPath() error = %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, KnowledgeDir, ErrorsFile)
	if path != expected {
		t.Errorf("GetErrorsPath() = %v, want %v", path, expected)
	}
}

func TestGetSummariesDir(t *testing.T) {
	path, err := GetSummariesDir()
	if err != nil {
		t.Fatalf("GetSummariesDir() error = %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, KnowledgeDir, SummariesDir)
	if path != expected {
		t.Errorf("GetSummariesDir() = %v, want %v", path, expected)
	}
}

func TestGetSummaryPath(t *testing.T) {
	date := "2026-04-06"
	path, err := GetSummaryPath(date)
	if err != nil {
		t.Fatalf("GetSummaryPath() error = %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, KnowledgeDir, SummariesDir, date+".md")
	if path != expected {
		t.Errorf("GetSummaryPath() = %v, want %v", path, expected)
	}
}

func TestEnsureDir(t *testing.T) {
	// 创建临时测试目录
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test-subdir")

	// 测试目录不存在时创建
	err := EnsureDir(testDir)
	if err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	// 验证目录已创建
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Errorf("EnsureDir() did not create directory %v", testDir)
	}

	// 测试目录已存在时不报错
	err = EnsureDir(testDir)
	if err != nil {
		t.Fatalf("EnsureDir() on existing dir error = %v", err)
	}
}
