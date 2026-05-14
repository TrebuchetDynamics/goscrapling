package spider

import (
	"context"
	"fmt"

	"github.com/TrebuchetDynamics/goscrapling"
)

type Session interface {
	Fetch(ctx context.Context, request Request) (*goscrapling.Response, error)
}

type Starter interface {
	Start(ctx context.Context) error
}

type Closer interface {
	Close(ctx context.Context) error
}

type SessionOptions struct {
	Default bool
	Lazy    bool
}

type SessionManager struct {
	sessions  map[string]Session
	lazy      map[string]bool
	started   map[string]bool
	defaultID string
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]Session),
		lazy:     make(map[string]bool),
		started:  make(map[string]bool),
	}
}

func (m *SessionManager) Add(id string, session Session, opts SessionOptions) error {
	if id == "" {
		return fmt.Errorf("session id is required")
	}
	if session == nil {
		return fmt.Errorf("session %q is nil", id)
	}
	if _, exists := m.sessions[id]; exists {
		return fmt.Errorf("session %q already registered", id)
	}

	m.sessions[id] = session
	m.lazy[id] = opts.Lazy
	if opts.Default || m.defaultID == "" {
		m.defaultID = id
	}
	return nil
}

func (m *SessionManager) Start(ctx context.Context) error {
	for id, session := range m.sessions {
		if m.lazy[id] {
			continue
		}
		if err := m.startSession(ctx, id, session); err != nil {
			return err
		}
	}
	return nil
}

func (m *SessionManager) Close(ctx context.Context) error {
	var firstErr error
	for id, session := range m.sessions {
		if !m.started[id] {
			continue
		}
		if closer, ok := session.(Closer); ok {
			if err := closer.Close(ctx); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		m.started[id] = false
	}
	return firstErr
}

func (m *SessionManager) Fetch(ctx context.Context, request Request) (Response, error) {
	id := request.SID
	if id == "" {
		id = m.defaultID
		request.SID = id
	}
	session, ok := m.sessions[id]
	if !ok {
		return Response{}, fmt.Errorf("session %q not found", id)
	}
	if err := m.startSession(ctx, id, session); err != nil {
		return Response{}, err
	}

	rawResponse, err := session.Fetch(ctx, request.clone())
	if err != nil {
		return Response{}, err
	}
	return Response{
		Response: rawResponse,
		Request:  request.clone(),
		Meta:     cloneMeta(request.Meta),
	}, nil
}

func (m *SessionManager) startSession(ctx context.Context, id string, session Session) error {
	if m.started[id] {
		return nil
	}
	if starter, ok := session.(Starter); ok {
		if err := starter.Start(ctx); err != nil {
			return err
		}
	}
	m.started[id] = true
	return nil
}
