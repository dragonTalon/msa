package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestManager_NewSession(t *testing.T) {
	m := GetManager()

	sess := m.NewSession(ModeTUI)

	if sess.UUID == "" {
		t.Error("NewSession() UUID should not be empty")
	}

	if sess.Mode != ModeTUI {
		t.Errorf("NewSession() Mode = %v, want %v", sess.Mode, ModeTUI)
	}

	if sess.CreatedAt.IsZero() {
		t.Error("NewSession() CreatedAt should not be zero")
	}

	if sess.FilePath == "" {
		t.Error("NewSession() FilePath should not be empty")
	}

	// FilePath should be in format: ~/.msa/memory/{YYYY-MM-DD_uuid}.md
	expectedPrefix := filepath.Join(m.memoryDir, sess.CreatedAt.Format("2006-01-02")+"_"+sess.UUID)
	if !strings.HasPrefix(sess.FilePath, expectedPrefix) {
		t.Errorf("NewSession() FilePath = %v, should start with %v", sess.FilePath, expectedPrefix)
	}
	if !strings.HasSuffix(sess.FilePath, ".md") {
		t.Errorf("NewSession() FilePath = %v, should end with .md", sess.FilePath)
	}
}

func TestManager_NewSession_CLI(t *testing.T) {
	m := GetManager()

	sess := m.NewSession(ModeCLI)

	if sess.Mode != ModeCLI {
		t.Errorf("NewSession() Mode = %v, want %v", sess.Mode, ModeCLI)
	}
}

func TestManager_CreateSessionFile(t *testing.T) {
	m := GetManager()

	// Use temp directory
	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	sess := m.NewSession(ModeTUI)

	err := m.CreateSessionFile(sess)
	if err != nil {
		t.Fatalf("CreateSessionFile() error = %v", err)
	}

	// Check file exists
	if _, err := os.Stat(sess.FilePath); os.IsNotExist(err) {
		t.Error("CreateSessionFile() did not create file")
	}

	// Read and verify content
	content, err := os.ReadFile(sess.FilePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	contentStr := string(content)

	// Check frontmatter
	if !strings.Contains(contentStr, "uuid: "+sess.UUID) {
		t.Error("CreateSessionFile() file should contain uuid")
	}
	if !strings.Contains(contentStr, "mode: tui") {
		t.Error("CreateSessionFile() file should contain mode")
	}
	if !strings.Contains(contentStr, "created_at:") {
		t.Error("CreateSessionFile() file should contain created_at")
	}
	if !strings.Contains(contentStr, "updated_at:") {
		t.Error("CreateSessionFile() file should contain updated_at")
	}
}

func TestManager_CreateSessionFile_FlatPath(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	sess := m.NewSession(ModeTUI)

	err := m.CreateSessionFile(sess)
	if err != nil {
		t.Fatalf("CreateSessionFile() error = %v", err)
	}

	// Check file exists at flat path (no subdirectories)
	if _, err := os.Stat(sess.FilePath); os.IsNotExist(err) {
		t.Error("CreateSessionFile() did not create file")
	}

	// Verify the file is directly in memoryDir
	expectedDir := tmpDir
	actualDir := filepath.Dir(sess.FilePath)
	if actualDir != expectedDir {
		t.Errorf("CreateSessionFile() file should be in %v, got %v", expectedDir, actualDir)
	}
}

func TestManager_AppendMessage_User(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	sess := m.NewSession(ModeTUI)

	// Create file first
	if err := m.CreateSessionFile(sess); err != nil {
		t.Fatalf("CreateSessionFile() error = %v", err)
	}

	err := m.AppendMessage(sess, "user", "Hello, this is a test message")
	if err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}

	// Read and verify
	content, err := os.ReadFile(sess.FilePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "## 👤 用户") {
		t.Error("AppendMessage() should contain user header")
	}
	if !strings.Contains(contentStr, "Hello, this is a test message") {
		t.Error("AppendMessage() should contain the message")
	}
}

func TestManager_AppendMessage_Assistant(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	sess := m.NewSession(ModeTUI)

	if err := m.CreateSessionFile(sess); err != nil {
		t.Fatalf("CreateSessionFile() error = %v", err)
	}

	err := m.AppendMessage(sess, "assistant", "This is the assistant reply")
	if err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}

	content, err := os.ReadFile(sess.FilePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "## 🤖 MSA") {
		t.Error("AppendMessage() should contain assistant header")
	}
	if !strings.Contains(contentStr, "This is the assistant reply") {
		t.Error("AppendMessage() should contain the reply")
	}
}

func TestManager_AppendMessage_NilSession(t *testing.T) {
	m := GetManager()

	err := m.AppendMessage(nil, "user", "test")
	if err != nil {
		t.Errorf("AppendMessage() with nil session should return nil, got %v", err)
	}
}

func TestManager_AppendMessage_EmptyPath(t *testing.T) {
	m := GetManager()

	sess := &Session{UUID: "test"}
	err := m.AppendMessage(sess, "user", "test")
	if err != nil {
		t.Errorf("AppendMessage() with empty path should return nil, got %v", err)
	}
}

func TestManager_AppendMessage_UnknownRole(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	sess := m.NewSession(ModeTUI)

	if err := m.CreateSessionFile(sess); err != nil {
		t.Fatalf("CreateSessionFile() error = %v", err)
	}

	err := m.AppendMessage(sess, "unknown", "test")
	if err != nil {
		t.Errorf("AppendMessage() with unknown role should return nil, got %v", err)
	}

	// File should not contain unknown role message
	content, _ := os.ReadFile(sess.FilePath)
	if strings.Contains(string(content), "unknown") {
		t.Error("AppendMessage() should not write unknown role")
	}
}

func TestManager_AppendMessage_AppendMode(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	sess := m.NewSession(ModeTUI)

	if err := m.CreateSessionFile(sess); err != nil {
		t.Fatalf("CreateSessionFile() error = %v", err)
	}

	// Append multiple messages
	m.AppendMessage(sess, "user", "Message 1")
	m.AppendMessage(sess, "assistant", "Reply 1")
	m.AppendMessage(sess, "user", "Message 2")

	content, err := os.ReadFile(sess.FilePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	contentStr := string(content)

	// All messages should be present
	if strings.Count(contentStr, "## 👤 用户") != 2 {
		t.Error("AppendMessage() should append user messages")
	}
	if strings.Count(contentStr, "## 🤖 MSA") != 1 {
		t.Error("AppendMessage() should append assistant messages")
	}
}

func TestFormatFrontmatter(t *testing.T) {
	sess := &Session{
		UUID:      "test-uuid-1234",
		CreatedAt: time.Date(2024, 3, 24, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 3, 24, 11, 0, 0, 0, time.UTC),
		Mode:      ModeTUI,
	}

	result := formatFrontmatter(sess)

	if !strings.Contains(result, "uuid: test-uuid-1234") {
		t.Error("formatFrontmatter() should contain uuid")
	}
	if !strings.Contains(result, "mode: tui") {
		t.Error("formatFrontmatter() should contain mode")
	}
	if !strings.Contains(result, "created_at: 2024-03-24T10:30:00Z") {
		t.Error("formatFrontmatter() should contain created_at")
	}
	if !strings.Contains(result, "updated_at: 2024-03-24T11:00:00Z") {
		t.Error("formatFrontmatter() should contain updated_at")
	}
	if !strings.HasPrefix(result, "---\n") {
		t.Error("formatFrontmatter() should start with ---")
	}
	if !strings.Contains(result, "\n---\n") {
		t.Error("formatFrontmatter() should contain closing ---")
	}
}
