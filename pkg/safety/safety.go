package safety

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"
)

type RollbackSession struct {
	ID             string
	DeviceName     string
	OriginalConfig string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	Timer          *time.Timer
	Done           chan struct{}
}

type SafetyEngine struct {
	mu       sync.Mutex
	sessions map[string]*RollbackSession
}

func NewSafetyEngine() *SafetyEngine {
	return &SafetyEngine{
		sessions: make(map[string]*RollbackSession),
	}
}

type SafetyWarning struct {
	Severity string `json:"severity" yaml:"severity"`
	Message  string `json:"message" yaml:"message"`
}

func (s *SafetyEngine) InspectCandidate(candidate string, mgmtIP string) []SafetyWarning {
	var warnings []SafetyWarning

	candLower := strings.ToLower(candidate)

	if mgmtIP != "" {
		if !strings.Contains(candidate, mgmtIP) {
			warnings = append(warnings, SafetyWarning{
				Severity: "high",
				Message:  fmt.Sprintf("Candidate config does not contain current management IP %s. Possible lockout risk.", mgmtIP),
			})
		}
	}

	if strings.Contains(candLower, "block all") || strings.Contains(candLower, "drop all") || strings.Contains(candLower, "default-action drop") {
		warnings = append(warnings, SafetyWarning{
			Severity: "medium",
			Message:  "Candidate config contains global drop/block rules.",
		})
	}

	if strings.Contains(candLower, "0.0.0.0/0") && strings.Contains(candLower, "delete") {
		warnings = append(warnings, SafetyWarning{
			Severity: "critical",
			Message:  "Possible deletion of default gateway route detected.",
		})
	}

	return warnings
}

func (s *SafetyEngine) StartCommitConfirm(
	id string,
	device string,
	origConfig string,
	timeoutSec int,
	onRollback func(id string, origConfig string) error,
) (*RollbackSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if timeoutSec <= 0 {
		timeoutSec = 60
	}

	session := &RollbackSession{
		ID:             id,
		DeviceName:     device,
		OriginalConfig: origConfig,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(time.Duration(timeoutSec) * time.Second),
		Done:           make(chan struct{}),
	}

	session.Timer = time.AfterFunc(time.Duration(timeoutSec)*time.Second, func() {
		s.mu.Lock()
		_, exists := s.sessions[id]
		if exists {
			delete(s.sessions, id)
		}
		s.mu.Unlock()

		if exists && onRollback != nil {
			_ = onRollback(id, origConfig)
		}
	})

	s.sessions[id] = session
	return session, nil
}

func (s *SafetyEngine) Confirm(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[id]
	if !exists {
		return errors.New("rollback session not found or already confirmed/expired")
	}

	if session.Timer != nil {
		session.Timer.Stop()
	}
	delete(s.sessions, id)
	return nil
}

func (s *SafetyEngine) ListPending() []*RollbackSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := make([]*RollbackSession, 0, len(s.sessions))
	for _, sess := range s.sessions {
		res = append(res, sess)
	}
	return res
}

func IsIPInPrefix(ipStr, prefixStr string) bool {
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return false
	}
	prefix, err := netip.ParsePrefix(prefixStr)
	if err != nil {
		return false
	}
	return prefix.Contains(ip)
}
