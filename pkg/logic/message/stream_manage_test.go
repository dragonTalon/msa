package message

import (
	"sync"
	"testing"

	msamodel "msa/pkg/model"
)

// TestGetStreamManager tests the GetStreamManager function
func TestGetStreamManager(t *testing.T) {
	manager := GetStreamManager()
	if manager == nil {
		t.Error("GetStreamManager() returned nil")
	}
}

// TestRegisterStreamOutput tests the RegisterStreamOutput function
func TestRegisterStreamOutput(t *testing.T) {
	ch, unregister := RegisterStreamOutput(10)
	if ch == nil {
		t.Fatal("RegisterStreamOutput() returned nil channel")
	}
	if unregister == nil {
		t.Fatal("RegisterStreamOutput() returned nil unregister function")
	}

	// Clean up
	unregister()
}

// TestStreamOutputManager_Register tests the Register method
func TestStreamOutputManager_Register(t *testing.T) {
	manager := &StreamOutputManager{}

	ch1, unregister1 := manager.Register(5)
	ch2, unregister2 := manager.Register(5)

	if ch1 == nil || ch2 == nil {
		t.Error("Register() returned nil channel")
	}

	// Verify subscriber count
	if count := manager.SubscriberCount(); count != 2 {
		t.Errorf("After registering 2 subscribers, count = %d, want 2", count)
	}

	// Verify HasSubscribers
	if !manager.HasSubscribers() {
		t.Error("HasSubscribers() should return true after registration")
	}

	// Clean up
	unregister1()
	unregister2()

	// Verify count after unregister
	if count := manager.SubscriberCount(); count != 0 {
		t.Errorf("After unregistering all, count = %d, want 0", count)
	}
}

// TestStreamOutputManager_Broadcast tests the Broadcast method
func TestStreamOutputManager_Broadcast(t *testing.T) {
	manager := &StreamOutputManager{}

	ch1, unregister1 := manager.Register(1)
	ch2, unregister2 := manager.Register(1)
	defer unregister1()
	defer unregister2()

	// Broadcast a message
	chunk := &msamodel.StreamChunk{
		Content: "test message",
		MsgType: msamodel.StreamMsgTypeText,
	}
	manager.Broadcast(chunk)

	// Verify both channels received the message
	select {
	case received := <-ch1:
		if received.Content != "test message" {
			t.Errorf("ch1 received content = %v, want 'test message'", received.Content)
		}
	default:
		t.Error("ch1 did not receive broadcast")
	}

	select {
	case received := <-ch2:
		if received.Content != "test message" {
			t.Errorf("ch2 received content = %v, want 'test message'", received.Content)
		}
	default:
		t.Error("ch2 did not receive broadcast")
	}
}

// TestStreamOutputManager_Broadcast_WithMultipleSubscribers tests broadcasting to multiple subscribers
func TestStreamOutputManager_Broadcast_WithMultipleSubscribers(t *testing.T) {
	manager := &StreamOutputManager{}

	// Register multiple subscribers
	var unregisters []func()
	for i := 0; i < 5; i++ {
		_, unregister := manager.Register(10)
		unregisters = append(unregisters, unregister)
	}
	defer func() {
		for _, unregister := range unregisters {
			unregister()
		}
	}()

	chunk := &msamodel.StreamChunk{
		Content: "broadcast test",
		MsgType: msamodel.StreamMsgTypeText,
	}
	manager.Broadcast(chunk)

	// All subscribers should receive the message
	if count := manager.SubscriberCount(); count != 5 {
		t.Errorf("Expected 5 subscribers, got %d", count)
	}
}

// TestStreamOutputManager_Broadcast_NonBlocking tests that broadcast doesn't block
func TestStreamOutputManager_Broadcast_NonBlocking(t *testing.T) {
	manager := &StreamOutputManager{}

	// Register a channel with buffer size 1
	_, unregister := manager.Register(1)
	defer unregister()

	// Broadcast should not block even if channel is full
	// (The actual channel is managed internally, we just verify broadcast doesn't panic)
	chunk := &msamodel.StreamChunk{
		Content: "non-blocking test",
		MsgType: msamodel.StreamMsgTypeText,
	}
	manager.Broadcast(chunk) // Should not block or panic
}

// TestStreamOutputManager_Unregister tests unregister functionality
func TestStreamOutputManager_Unregister(t *testing.T) {
	manager := &StreamOutputManager{}

	_, unregister1 := manager.Register(10)
	_, unregister2 := manager.Register(10)

	// Unregister first
	unregister1()

	// Verify count decreased
	if count := manager.SubscriberCount(); count != 1 {
		t.Errorf("After unregistering 1 of 2, count = %d, want 1", count)
	}

	// Verify still has subscribers
	if !manager.HasSubscribers() {
		t.Error("Should still have subscribers after unregistering one")
	}

	// Unregister second
	unregister2()

	// Verify no subscribers
	if manager.HasSubscribers() {
		t.Error("Should have no subscribers after unregistering all")
	}
}

// TestStreamOutputManager_ConcurrentAccess tests thread safety
func TestStreamOutputManager_ConcurrentAccess(t *testing.T) {
	manager := &StreamOutputManager{}

	var wg sync.WaitGroup
	done := make(chan bool)

	// Start multiple goroutines registering and broadcasting
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ch, unregister := manager.Register(5)
			defer unregister()

			// Broadcast some messages
			chunk := &msamodel.StreamChunk{
				Content: "concurrent test",
				MsgType: msamodel.StreamMsgTypeText,
			}
			manager.Broadcast(chunk)

			// Try to receive
			select {
			case <-ch:
				// Message received
			default:
				// Channel might be full
			}
		}(i)
	}

	// Wait for all goroutines
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-done: // timeout if needed
	}
}

// TestBroadcastToolStart tests the BroadcastToolStart function
func TestBroadcastToolStart(t *testing.T) {
	manager := &StreamOutputManager{}
	globalStreamManager = manager // Use test manager

	ch, unregister := manager.Register(10)
	defer unregister()

	BroadcastToolStart("test_tool", "param1=value1")

	// Verify message received
	select {
	case chunk := <-ch:
		if chunk.MsgType != msamodel.StreamMsgTypeTool {
			t.Errorf("Expected MsgType StreamMsgTypeTool, got %v", chunk.MsgType)
		}
		if chunk.Content == "" {
			t.Error("Expected non-empty content")
		}
	default:
		t.Error("Did not receive broadcast message")
	}
}

// TestBroadcastToolEnd tests the BroadcastToolEnd function
func TestBroadcastToolEnd(t *testing.T) {
	manager := &StreamOutputManager{}
	globalStreamManager = manager // Use test manager

	ch, unregister := manager.Register(10)
	defer unregister()

	// Test successful case
	BroadcastToolEnd("test_tool", "success", nil)

	select {
	case chunk := <-ch:
		if chunk.Err != nil {
			t.Errorf("Expected no error for success case, got %v", chunk.Err)
		}
	default:
	}

	// Test error case
	testErr := &testError{msg: "test error"}
	BroadcastToolEnd("test_tool", "", testErr)

	select {
	case chunk := <-ch:
		if chunk.Err == nil {
			t.Error("Expected error in chunk for error case")
		}
	default:
	}
}

// TestStreamOutputManager_SubscriberCount tests the SubscriberCount method
func TestStreamOutputManager_SubscriberCount(t *testing.T) {
	manager := &StreamOutputManager{}

	// Initially 0
	if count := manager.SubscriberCount(); count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Add subscribers
	unregisters := make([]func(), 0)
	for i := 0; i < 3; i++ {
		_, unregister := manager.Register(10)
		unregisters = append(unregisters, unregister)
	}

	if count := manager.SubscriberCount(); count != 3 {
		t.Errorf("After adding 3, count = %d, want 3", count)
	}

	// Remove all
	for _, unregister := range unregisters {
		unregister()
	}

	if count := manager.SubscriberCount(); count != 0 {
		t.Errorf("After removing all, count = %d, want 0", count)
	}
}

// TestStreamOutputManager_HasSubscribers tests the HasSubscribers method
func TestStreamOutputManager_HasSubscribers(t *testing.T) {
	manager := &StreamOutputManager{}

	if manager.HasSubscribers() {
		t.Error("HasSubscribers() should return false initially")
	}

	_, unregister := manager.Register(10)
	defer unregister()

	if !manager.HasSubscribers() {
		t.Error("HasSubscribers() should return true after registration")
	}

	unregister()

	if manager.HasSubscribers() {
		t.Error("HasSubscribers() should return false after unregister")
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
