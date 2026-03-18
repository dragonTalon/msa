package memory

import (
	"testing"
	"time"
)

// TestGetAutoTag 测试自动标签函数
func TestGetAutoTag(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string // 格式：15:04
		expected string
	}{
		// 早盘时段
		{"早盘开始 9:30", "09:30", TagMorningSession},
		{"早盘中间 10:00", "10:00", TagMorningSession},
		{"早盘结束 11:30", "11:30", TagMorningSession},
		{"早盘 10:15", "10:15", TagMorningSession},

		// 午盘时段
		{"午盘开始 13:00", "13:00", TagAfternoonSession},
		{"午盘中间 14:00", "14:00", TagAfternoonSession},
		{"午盘结束 14:30", "14:30", TagAfternoonSession},
		{"午盘 13:30", "13:30", TagAfternoonSession},

		// 收盘时段
		{"收盘开始 16:00", "16:00", TagCloseSession},
		{"收盘后 17:00", "17:00", TagCloseSession},
		{"收盘后 20:00", "20:00", TagCloseSession},
		{"深夜 23:00", "23:00", TagCloseSession},

		// 非交易时段
		{"开盘前 8:00", "08:00", ""},
		{"午休前 11:45", "11:45", ""},
		{"午休中 12:00", "12:00", ""},
		{"午休结束 12:45", "12:45", ""},
		{"收盘前 15:00", "15:00", ""},
		{"收盘前 15:30", "15:30", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 解析时间字符串
			parsedTime, err := time.Parse("15:04", tt.timeStr)
			if err != nil {
				t.Fatalf("解析时间失败: %v", err)
			}

			// 使用今天的日期 + 指定时间
			now := time.Now()
			testTime := time.Date(now.Year(), now.Month(), now.Day(),
				parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())

			result := getAutoTag(testTime)
			if result != tt.expected {
				t.Errorf("getAutoTag(%s) = %q, want %q", tt.timeStr, result, tt.expected)
			}
		})
	}
}

// TestGetAutoTagBoundaries 测试边界时间
func TestGetAutoTagBoundaries(t *testing.T) {
	now := time.Now()

	// 边界时间测试
	boundaryTests := []struct {
		timeStr  string
		hour     int
		minute   int
		expected string
	}{
		// 早盘边界
		{"09:29", 9, 29, ""},                 // 早盘前1分钟
		{"09:30", 9, 30, TagMorningSession},  // 早盘开始
		{"11:30", 11, 30, TagMorningSession}, // 早盘结束
		{"11:31", 11, 31, ""},                // 早盘后1分钟

		// 午盘边界
		{"12:59", 12, 59, ""},                  // 午盘前1分钟
		{"13:00", 13, 0, TagAfternoonSession},  // 午盘开始
		{"14:30", 14, 30, TagAfternoonSession}, // 午盘结束
		{"14:31", 14, 31, ""},                  // 午盘后1分钟

		// 收盘边界
		{"15:59", 15, 59, ""},             // 收盘前1分钟
		{"16:00", 16, 0, TagCloseSession}, // 收盘开始
	}

	for _, tt := range boundaryTests {
		t.Run(tt.timeStr, func(t *testing.T) {
			testTime := time.Date(now.Year(), now.Month(), now.Day(),
				tt.hour, tt.minute, 0, 0, now.Location())

			result := getAutoTag(testTime)
			if result != tt.expected {
				t.Errorf("getAutoTag(%s) = %q, want %q", tt.timeStr, result, tt.expected)
			}
		})
	}
}
