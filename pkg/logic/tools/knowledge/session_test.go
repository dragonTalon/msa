package knowledge

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(time.Time) bool
	}{
		{
			name:  "today",
			input: "today",
			checkFunc: func(got time.Time) bool {
				now := time.Now()
				return got.Year() == now.Year() && got.Month() == now.Month() && got.Day() == now.Day()
			},
		},
		{
			name:  "yesterday",
			input: "yesterday",
			checkFunc: func(got time.Time) bool {
				yesterday := time.Now().AddDate(0, 0, -1)
				return got.Year() == yesterday.Year() && got.Month() == yesterday.Month() && got.Day() == yesterday.Day()
			},
		},
		{
			name:    "ISO格式日期",
			input:   "2026-03-15",
			wantErr: false,
			checkFunc: func(got time.Time) bool {
				return got.Year() == 2026 && got.Month() == 3 && got.Day() == 15
			},
		},
		{
			name:    "TODAY大写",
			input:   "TODAY",
			wantErr: false,
			checkFunc: func(got time.Time) bool {
				now := time.Now()
				return got.Year() == now.Year() && got.Month() == now.Month() && got.Day() == now.Day()
			},
		},
		{
			name:    "无效日期格式",
			input:   "invalid-date",
			wantErr: true,
		},
		{
			name:    "空字符串",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDate(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("parseDate(%q) unexpected error: %v", tt.input, err)
				return
			}

			if tt.checkFunc != nil && !tt.checkFunc(got) {
				t.Errorf("parseDate(%q) = %v, check failed", tt.input, got)
			}
		})
	}
}

func TestIsSameDate(t *testing.T) {
	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		expected bool
	}{
		{
			name:     "相同日期",
			t1:       time.Date(2026, 3, 15, 10, 30, 0, 0, time.Local),
			t2:       time.Date(2026, 3, 15, 15, 45, 0, 0, time.Local),
			expected: true,
		},
		{
			name:     "不同日期",
			t1:       time.Date(2026, 3, 15, 10, 30, 0, 0, time.Local),
			t2:       time.Date(2026, 3, 16, 10, 30, 0, 0, time.Local),
			expected: false,
		},
		{
			name:     "不同月份",
			t1:       time.Date(2026, 3, 15, 10, 30, 0, 0, time.Local),
			t2:       time.Date(2026, 4, 15, 10, 30, 0, 0, time.Local),
			expected: false,
		},
		{
			name:     "不同年份",
			t1:       time.Date(2026, 3, 15, 10, 30, 0, 0, time.Local),
			t2:       time.Date(2025, 3, 15, 10, 30, 0, 0, time.Local),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameDate(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("isSameDate(%v, %v) = %v, want %v", tt.t1, tt.t2, result, tt.expected)
			}
		})
	}
}

func TestContainsTag(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		tag      string
		expected bool
	}{
		{
			name:     "标签存在",
			tags:     []string{"morning-session", "trading"},
			tag:      "morning-session",
			expected: true,
		},
		{
			name:     "标签不存在",
			tags:     []string{"morning-session", "trading"},
			tag:      "afternoon-session",
			expected: false,
		},
		{
			name:     "空标签列表",
			tags:     []string{},
			tag:      "morning-session",
			expected: false,
		},
		{
			name:     "大小写不敏感",
			tags:     []string{"Morning-Session", "Trading"},
			tag:      "morning-session",
			expected: true,
		},
		{
			name:     "标签大写查找",
			tags:     []string{"morning-session"},
			tag:      "MORNING-SESSION",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsTag(tt.tags, tt.tag)
			if result != tt.expected {
				t.Errorf("containsTag(%v, %q) = %v, want %v", tt.tags, tt.tag, result, tt.expected)
			}
		})
	}
}

func TestValidateTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantErr bool
	}{
		{
			name:    "有效标签-小写字母",
			tag:     "morning-session",
			wantErr: false,
		},
		{
			name:    "有效标签-包含数字",
			tag:     "session-2026",
			wantErr: false,
		},
		{
			name:    "有效标签-纯字母",
			tag:     "trading",
			wantErr: false,
		},
		{
			name:    "有效标签-大写字母",
			tag:     "Morning-Session",
			wantErr: false,
		},
		{
			name:    "无效标签-空字符串",
			tag:     "",
			wantErr: true,
		},
		{
			name:    "无效标签-包含空格",
			tag:     "morning session",
			wantErr: true,
		},
		{
			name:    "无效标签-包含特殊字符",
			tag:     "morning@session",
			wantErr: true,
		},
		{
			name:    "无效标签-包含下划线",
			tag:     "morning_session",
			wantErr: true,
		},
		{
			name:    "无效标签-中文",
			tag:     "早盘会话",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTag(tt.tag)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateTag(%q) expected error, got nil", tt.tag)
				}
				return
			}

			if err != nil {
				t.Errorf("validateTag(%q) unexpected error: %v", tt.tag, err)
			}
		})
	}
}

func TestFormatSessionsResult(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		sessions []interface{} // 简化测试，使用空列表
		tag      string
		contains []string
	}{
		{
			name:     "空会话列表",
			date:     time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local),
			sessions: nil,
			tag:      "",
			contains: []string{"2026-03-15", "未找到符合条件的会话"},
		},
		{
			name:     "带标签过滤",
			date:     time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local),
			sessions: nil,
			tag:      "morning-session",
			contains: []string{"2026-03-15", "标签过滤", "morning-session"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSessionsResult(tt.date, nil, tt.tag)

			for _, s := range tt.contains {
				if !containsString(result, s) {
					t.Errorf("formatSessionsResult() result missing %q", s)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s[1:], substr) || len(s) >= len(substr) && s[:len(substr)] == substr)
}
