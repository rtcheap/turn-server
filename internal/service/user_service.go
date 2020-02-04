package service

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/CzarSimon/httputil"
	"github.com/pion/turn/v2"
	"github.com/rtcheap/dto"
	"github.com/rtcheap/turn-server/internal/repository"
)

// UserService handles adding, key generation and retrieving users.
type UserService struct {
	startedSessions uint64
	endedSessions   uint64
	Realm           string
	Keys            repository.KeyRepository
}

// FindKey looks up and retuns the key for a given username.
func (s *UserService) FindKey(username, realm string, addr net.Addr) ([]byte, bool) {
	return s.Keys.Find(username)
}

// CreateKey creates a key from a user session and stores it for future use.
func (s *UserService) CreateKey(session dto.Session) error {
	if _, ok := s.Keys.Find(session.UserID); ok {
		err := fmt.Errorf("session already exists %s", session)
		return httputil.ConflictError(err)
	}

	key := turn.GenerateAuthKey(session.UserID, s.Realm, session.Key)
	err := s.Keys.Save(session.UserID, key)
	if err != nil {
		return fmt.Errorf("failed to store session key created for %s. %w", session, err)
	}

	atomic.AddUint64(&s.startedSessions, 1)
	return nil
}

// GetStatistics returns the gathered
func (s *UserService) GetStatistics() dto.SessionStatistics {
	return dto.SessionStatistics{
		Started: s.startedSessions,
		Ended:   s.endedSessions,
	}
}
