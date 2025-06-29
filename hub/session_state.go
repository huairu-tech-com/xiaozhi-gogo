package hub

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type SessionStateKind string

const (
	kSessionStateIdle       SessionStateKind = "idle"
	kSessionStateConnecting SessionStateKind = "connecting"
	kSessionStateListening  SessionStateKind = "listening"
	kSessionStateSpeaking   SessionStateKind = "speaking"
)

type TransitionCallback func(s *Session, from SessionStateKind, to SessionStateKind) error
type KindPair [2]SessionStateKind

type SessionState struct {
	s    *Session         // session is used to access session properties
	kind SessionStateKind // kind can not be assign directly, use methods to change it

	// the following transitions are valid for this state
	ValidTransitions map[SessionStateKind][]SessionStateKind
	Callbacks        map[KindPair][]TransitionCallback // callbacks for transitions, key is [from, to] pair
}

func newSessionState(s *Session, kind SessionStateKind) *SessionState {
	return &SessionState{
		s:                s,
		kind:             kind,
		ValidTransitions: make(map[SessionStateKind][]SessionStateKind),
		Callbacks:        make(map[KindPair][]TransitionCallback),
	}
}

func (s *SessionState) IsValidTransition(to SessionStateKind) bool {
	validTransitions, ok := s.ValidTransitions[s.kind]
	if !ok {
		return false
	}

	return lo.Contains(validTransitions, to)
}

func (s *SessionState) TransitTo(newState SessionStateKind) error {
	if !s.IsValidTransition(newState) {
		return errors.Errorf("invalid transition from %s to %s", s.kind, newState)
	}

	s.kind = newState
	callbacks, ok := s.Callbacks[KindPair{s.kind, newState}]
	if !ok {
		return nil
	}

	var err error
	for _, callback := range callbacks {
		callbackErr := callback(s.s, s.kind, newState)
		err = errors.Wrapf(callbackErr, "callback for transition from %s to %s failed", s.kind, newState)
	}

	return err
}

func logTransition(s *Session, from SessionStateKind, to SessionStateKind) error {
	log.Info().Msgf("Session %s transitioned from %s to %s", s.sessionId, from, to)
	return nil
}
