package session

import (
	"sync"
	"testing"
)

func TestGetManager_Singleton(t *testing.T) {
	m1 := GetManager()
	m2 := GetManager()

	if m1 != m2 {
		t.Error("GetManager() should return the same instance")
	}
}

func TestGetManager_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	managers := make([]*Manager, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			managers[idx] = GetManager()
		}(i)
	}

	wg.Wait()

	// All should be the same instance
	first := managers[0]
	for i, m := range managers {
		if m != first {
			t.Errorf("managers[%d] != managers[0], singleton not working", i)
		}
	}
}

func TestManager_Current(t *testing.T) {
	m := GetManager()
	defer m.Clear()

	// Initially nil
	if m.Current() != nil {
		t.Error("Current() should be nil initially")
	}

	// Set a session
	sess := &Session{UUID: "test-uuid"}
	m.SetCurrent(sess)

	if m.Current() != sess {
		t.Error("Current() should return the set session")
	}
}

func TestManager_Clear(t *testing.T) {
	m := GetManager()

	sess := &Session{UUID: "test-uuid"}
	m.SetCurrent(sess)

	if m.Current() == nil {
		t.Error("Current() should not be nil after SetCurrent")
	}

	m.Clear()

	if m.Current() != nil {
		t.Error("Current() should be nil after Clear")
	}
}

func TestManager_GetMemoryDir(t *testing.T) {
	m := GetManager()

	dir := m.GetMemoryDir()
	if dir == "" {
		t.Error("GetMemoryDir() should not return empty string")
	}

	// Should contain .msa/memory
	if !contains(dir, ".msa") || !contains(dir, "memory") {
		t.Errorf("GetMemoryDir() = %v, should contain .msa/memory", dir)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
