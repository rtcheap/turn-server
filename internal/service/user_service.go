package service

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/CzarSimon/httputil"
	"github.com/pion/turn/v2"
	"github.com/rtcheap/turn-server/internal/repository"
)

// Session user session.
type Session struct {
	UserID string `json:"userId,omitempty"`
	Key    string `json:"key,omitempty"`
}

// SessionStatistics statistics for the numbers of handled sessions
type SessionStatistics struct {
	Count uint64 `json:"count"`
}

// UserService handles adding, key generation and retrieving users.
type UserService struct {
	sessionCounter uint64
	Realm          string
	Keys           repository.KeyRepository
}

// FindKey looks up and retuns the key for a given username.
func (s *UserService) FindKey(username, realm string, addr net.Addr) ([]byte, bool) {
	return s.Keys.Find(username)
}

// CreateKey creates a key from a user session and stores it for future use.
func (s *UserService) CreateKey(session Session) error {
	if _, ok := s.Keys.Find(session.UserID); ok {
		err := fmt.Errorf("session already exists for user(id=%s)", session.UserID)
		return httputil.ConflictError(err)
	}

	key := turn.GenerateAuthKey(session.UserID, s.Realm, session.Key)
	err := s.Keys.Save(session.UserID, key)
	if err != nil {
		return fmt.Errorf("failed to store session key created for user(id=%s). %w", session.UserID, err)
	}

	atomic.AddUint64(&s.sessionCounter, 1)
	return nil
}

// GetStatistics returns the gathered
func (s *UserService) GetStatistics() SessionStatistics {
	return SessionStatistics{
		Count: s.sessionCounter,
	}
}
