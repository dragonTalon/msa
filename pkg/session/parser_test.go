package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseError_Error(t *testing.T) {
	err := &ParseError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("ParseError.Error() = %v, want %v", err.Error(), "test error")
	}
}

func TestManager_LoadSession(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	// Create a test session file with flat path
	sess := &Session{
		UUID:      "a1b2c3d4-5678-90ab-cdef-1234567890ab",
		CreatedAt: time.Date(2024, 3, 24, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 3, 24, 11, 0, 0, 0, time.UTC),
		Mode:      ModeTUI,
	}

	// Create file with flat path: {YYYY-MM-DD_uuid}.md
	fileName := fmt.Sprintf("2024-03-24_%s.md", sess.UUID)
	filePath := filepath.Join(tmpDir, fileName)
	sess.FilePath = filePath

	// Write test content
	content := `---
uuid: a1b2c3d4-5678-90ab-cdef-1234567890ab
created_at: 2024-03-24T10:30:00Z
updated_at: 2024-03-24T11:00:00Z
mode: tui
---

## 👤 用户
Hello, how are you?

## 🤖 MSA
I am fine, thank you!
`
	os.WriteFile(filePath, []byte(content), 0644)

	// Test LoadSession
	parsed, err := m.LoadSession("2024-03-24_a1b2c3d4")
	if err != nil {
		t.Fatalf("LoadSession() error = %v", err)
	}

	if parsed.Session.UUID != sess.UUID {
		t.Errorf("LoadSession() UUID = %v, want %v", parsed.Session.UUID, sess.UUID)
	}

	if len(parsed.Messages) != 2 {
		t.Errorf("LoadSession() Messages count = %v, want 2", len(parsed.Messages))
	}
}

func TestManager_LoadSession_NotFound(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	_, err := m.LoadSession("2024-03-24_notexist")
	if err == nil {
		t.Error("LoadSession() should return error for non-existent session")
	}
}

func TestManager_LoadSession_InvalidFormat(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	// Create file with flat path
	filePath := filepath.Join(tmpDir, "2024-03-24_test-uuid.md")

	// Invalid format - no frontmatter
	content := "This is not a valid session file"
	os.WriteFile(filePath, []byte(content), 0644)

	_, err := m.LoadSession("2024-03-24_test-uuid")
	if err == nil {
		t.Error("LoadSession() should return error for invalid format")
	}
}

func TestParseSessionID(t *testing.T) {
	m := GetManager()

	tmpDir := t.TempDir()
	m.memoryDir = tmpDir

	// Create test file with flat path
	testUUID := "a1b2c3d4-5678-90ab-cdef-1234567890ab"
	fileName := fmt.Sprintf("2024-03-24_%s.md", testUUID)
	filePath := filepath.Join(tmpDir, fileName)
	os.WriteFile(filePath, []byte("test"), 0644)

	tests := []struct {
		name       string
		sessionID  string
		wantErr    bool
		errContain string
	}{
		{
			name:      "valid short id",
			sessionID: "2024-03-24_a1b2c3d4",
			wantErr:   false,
		},
		{
			name:       "invalid format - no underscore",
			sessionID:  "2024-03-24",
			wantErr:    true,
			errContain: "格式错误",
		},
		{
			name:       "invalid date format",
			sessionID:  "invalid_uuid",
			wantErr:    true,
			errContain: "日期格式错误",
		},
		{
			name:       "non-existent session",
			sessionID:  "2024-03-24_notexist",
			wantErr:    true,
			errContain: "不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := m.parseSessionID(tt.sessionID)
			if tt.wantErr {
				if err == nil {
					t.Error("parseSessionID() should return error")
				} else if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("parseSessionID() error = %v, should contain %v", err, tt.errContain)
				}
			} else {
				if err != nil {
					t.Errorf("parseSessionID() error = %v", err)
				}
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantUUID string
		wantMode Mode
		wantErr  bool
	}{
		{
			name: "valid frontmatter",
			input: `uuid: test-uuid-123
created_at: 2024-03-24T10:30:00Z
updated_at: 2024-03-24T11:00:00Z
mode: cli`,
			wantUUID: "test-uuid-123",
			wantMode: ModeCLI,
			wantErr:  false,
		},
		{
			name: "missing uuid",
			input: `created_at: 2024-03-24T10:30:00Z
mode: tui`,
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess, err := parseFrontmatter(tt.input, "/tmp/test.md")

			if tt.wantErr {
				if err == nil {
					t.Error("parseFrontmatter() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseFrontmatter() error = %v", err)
			}

			if sess.UUID != tt.wantUUID {
				t.Errorf("parseFrontmatter() UUID = %v, want %v", sess.UUID, tt.wantUUID)
			}

			if sess.Mode != tt.wantMode {
				t.Errorf("parseFrontmatter() Mode = %v, want %v", sess.Mode, tt.wantMode)
			}
		})
	}
}

func TestParseMessages(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantCount     int
		wantFirstRole string
		wantLastRole  string
	}{
		{
			name: "single user message",
			input: `## 👤 用户
Hello, how are you?
`,
			wantCount:     1,
			wantFirstRole: "user",
		},
		{
			name: "user and assistant",
			input: `## 👤 用户
Hello!

## 🤖 MSA
Hi there!
`,
			wantCount:     2,
			wantFirstRole: "user",
			wantLastRole:  "assistant",
		},
		{
			name: "multiple rounds",
			input: `## 👤 用户
Question 1

## 🤖 MSA
Answer 1

## 👤 用户
Question 2

## 🤖 MSA
Answer 2
`,
			wantCount:     4,
			wantFirstRole: "user",
			wantLastRole:  "assistant",
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
		{
			name: "ignore other content",
			input: `## Other Header
This should be ignored

## 👤 用户
Real message
`,
			wantCount:     1,
			wantFirstRole: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := parseMessages(tt.input)

			if len(messages) != tt.wantCount {
				t.Errorf("parseMessages() count = %v, want %v", len(messages), tt.wantCount)
			}

			if tt.wantCount > 0 && string(messages[0].Role) != tt.wantFirstRole {
				t.Errorf("parseMessages() first role = %v, want %v", messages[0].Role, tt.wantFirstRole)
			}

			if tt.wantCount > 1 && string(messages[len(messages)-1].Role) != tt.wantLastRole {
				t.Errorf("parseMessages() last role = %v, want %v", messages[len(messages)-1].Role, tt.wantLastRole)
			}
		})
	}
}

func TestParseSessionFile(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantErr    bool
		errContain string
	}{
		{
			name: "valid file",
			content: `---
uuid: test-uuid
created_at: 2024-03-24T10:30:00Z
updated_at: 2024-03-24T11:00:00Z
mode: tui
---

## 👤 用户
Hello
`,
			wantErr: false,
		},
		{
			name:       "missing frontmatter",
			content:    "No frontmatter here",
			wantErr:    true,
			errContain: "缺少元数据",
		},
		{
			name: "unclosed frontmatter",
			content: `---
uuid: test-uuid
mode: tui

## 👤 用户
Hello
`,
			wantErr:    true,
			errContain: "格式错误",
		},
		{
			name: "missing uuid in frontmatter",
			content: `---
created_at: 2024-03-24T10:30:00Z
mode: tui
---

## 👤 用户
Hello
`,
			wantErr:    true,
			errContain: "缺少 uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSessionFile([]byte(tt.content), "/tmp/test.md")

			if tt.wantErr {
				if err == nil {
					t.Error("parseSessionFile() should return error")
				} else if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("parseSessionFile() error = %v, should contain %v", err, tt.errContain)
				}
			} else {
				if err != nil {
					t.Errorf("parseSessionFile() error = %v", err)
				}
			}
		})
	}
}
