package internal

import (
	"fmt"
	"sync"
	"time"
)

// HealthState 引擎健康状态
type HealthState string

const (
	StateHealthy   HealthState = "healthy"
	StateDegraded  HealthState = "degraded"
	StateUnhealthy HealthState = "unhealthy"
)

// EngineHealth 引擎健康状态
type EngineHealth struct {
	Name          string
	State         HealthState
	FailureCount  int
	LastFailTime  time.Time
	CooldownUntil time.Time
}

// EngineTracker 引擎健康追踪器
type EngineTracker struct {
	mu               sync.RWMutex
	health           map[string]*EngineHealth
	cooldown         time.Duration // 冷却期时长
	failureThreshold int           // 触发熔断的失败次数
}

// NewEngineTracker 创建引擎追踪器
func NewEngineTracker() *EngineTracker {
	return &EngineTracker{
		health:           make(map[string]*EngineHealth),
		cooldown:         5 * time.Minute, // 默认 5 分钟
		failureThreshold: 3,               // 默认 3 次失败触发熔断
	}
}

// SetCooldown 设置冷却期时长
func (t *EngineTracker) SetCooldown(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cooldown = d
}

// SetFailureThreshold 设置触发熔断的失败次数
func (t *EngineTracker) SetFailureThreshold(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.failureThreshold = n
}

// ShouldTry 判断是否应该尝试该引擎
func (t *EngineTracker) ShouldTry(engine string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	h, exists := t.health[engine]
	if !exists {
		return true // 新引擎，默认可用
	}

	if h.State == StateUnhealthy {
		if time.Now().Before(h.CooldownUntil) {
			return false // 仍在冷却期内
		}
		// 冷却期已过，尝试恢复
		t.mu.RUnlock()
		t.mu.Lock()
		h.State = StateDegraded
		h.FailureCount = 0
		t.mu.Unlock()
		t.mu.RLock()
	}

	return true
}

// RecordFailure 记录引擎失败
func (t *EngineTracker) RecordFailure(engine string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	h, exists := t.health[engine]
	if !exists {
		h = &EngineHealth{
			Name:  engine,
			State: StateHealthy,
		}
		t.health[engine] = h
	}

	h.FailureCount++
	h.LastFailTime = time.Now()

	// 达到失败阈值，标记为不健康
	if h.FailureCount >= t.failureThreshold {
		h.State = StateUnhealthy
		h.CooldownUntil = time.Now().Add(t.cooldown)
	}

	// 如果是 CAPTCHA 错误，立即标记为不健康
	if err != nil && isCaptchaError(err) {
		h.State = StateUnhealthy
		h.CooldownUntil = time.Now().Add(t.cooldown)
	}
}

// RecordSuccess 记录引擎成功
func (t *EngineTracker) RecordSuccess(engine string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	h, exists := t.health[engine]
	if !exists {
		h = &EngineHealth{
			Name:  engine,
			State: StateHealthy,
		}
		t.health[engine] = h
		return
	}

	// 成功后恢复健康状态
	if h.State == StateDegraded {
		h.State = StateHealthy
		h.FailureCount = 0
	}
}

// GetHealth 获取引擎健康状态
func (t *EngineTracker) GetHealth(engine string) *EngineHealth {
	t.mu.RLock()
	defer t.mu.RUnlock()

	h, exists := t.health[engine]
	if !exists {
		return &EngineHealth{
			Name:  engine,
			State: StateHealthy,
		}
	}

	// 返回副本
	return &EngineHealth{
		Name:          h.Name,
		State:         h.State,
		FailureCount:  h.FailureCount,
		LastFailTime:  h.LastFailTime,
		CooldownUntil: h.CooldownUntil,
	}
}

// GetState 获取引擎状态（用于日志）
func (t *EngineTracker) GetState(engine string) HealthState {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if h, exists := t.health[engine]; exists {
		return h.State
	}
	return StateHealthy
}

// isCaptchaError 判断是否为 CAPTCHA 错误
func isCaptchaError(err error) bool {
	if err == nil {
		return false
	}
	errStr := fmt.Sprintf("%v", err)
	captchaIndicators := []string{
		"captcha",
		"unusual traffic",
		"请证明您不是机器人",
		"verify you are not a robot",
	}
	lowerErrStr := toLower(errStr)
	for _, indicator := range captchaIndicators {
		if contains(lowerErrStr, toLower(indicator)) {
			return true
		}
	}
	return false
}

// toLower 转换为小写（避免使用 strings 包以减少依赖）
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring 查找子串位置
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
