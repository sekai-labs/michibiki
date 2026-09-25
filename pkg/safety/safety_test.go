package safety

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSafetyInspection(t *testing.T) {
	eng := NewSafetyEngine()
	warnings := eng.InspectCandidate("interface eth0\n ip address 10.0.0.1/24\n drop all", "192.168.1.1")
	if len(warnings) == 0 {
		t.Fatalf("expected warnings, got none")
	}
}

func TestCommitConfirmRollbackTrigger(t *testing.T) {
	eng := NewSafetyEngine()

	var rolledBack atomic.Bool
	_, err := eng.StartCommitConfirm("sess-1", "test-router", "orig-config", 1, func(id, orig string) error {
		rolledBack.Store(true)
		return nil
	})
	if err != nil {
		t.Fatalf("StartCommitConfirm failed: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	if !rolledBack.Load() {
		t.Fatalf("expected rollback to be triggered")
	}
}

func TestCommitConfirmSuccess(t *testing.T) {
	eng := NewSafetyEngine()

	var rolledBack atomic.Bool
	_, err := eng.StartCommitConfirm("sess-2", "test-router", "orig-config", 2, func(id, orig string) error {
		rolledBack.Store(true)
		return nil
	})
	if err != nil {
		t.Fatalf("StartCommitConfirm failed: %v", err)
	}

	if err := eng.Confirm("sess-2"); err != nil {
		t.Fatalf("Confirm failed: %v", err)
	}

	time.Sleep(2200 * time.Millisecond)

	if rolledBack.Load() {
		t.Fatalf("rollback should not have been triggered after confirmation")
	}
}
