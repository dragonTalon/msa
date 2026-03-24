package session

import (
	"testing"
	"time"
)

func TestSession_ShortID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected string
	}{
		{
			name:     "normal uuid",
			uuid:     "a1b2c3d4-5678-90ab-cdef-1234567890ab",
			expected: "a1b2c3d4",
		},
		{
			name:     "short uuid",
			uuid:     "abc",
			expected: "abc",
		},
		{
			name:     "empty uuid",
			uuid:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{UUID: tt.uuid}
			if got := s.ShortID(); got != tt.expected {
				t.Errorf("ShortID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSession_SessionID(t *testing.T) {
	created := time.Date(2024, 3, 24, 10, 30, 0, 0, time.Local)
	s := &Session{
		UUID:      "a1b2c3d4-5678-90ab-cdef-1234567890ab",
		CreatedAt: created,
	}

	expected := "2024-03-24_a1b2c3d4"
	if got := s.SessionID(); got != expected {
		t.Errorf("SessionID() = %v, want %v", got, expected)
	}
}

func TestMode_String(t *testing.T) {
	tests := []struct {
		mode     Mode
		expected string
	}{
		{ModeTUI, "tui"},
		{ModeCLI, "cli"},
		{Mode("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("Mode = %v, want %v", tt.mode, tt.expected)
			}
		})
	}
}
