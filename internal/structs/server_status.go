package structs

import "sync"

// ServerStatus represents status of a server. It is concurrency safe.
type ServerStatus struct {
	mu  sync.Mutex
	val ServerStatusValue
}

// NewServerStatus returns a new status instance given an initial value.
func NewServerStatus(v ServerStatusValue) *ServerStatus {
	return &ServerStatus{val: v}
}

type ServerStatusValue int

const (
	// StatusIdle indicates the server is in idle state.
	StatusIdle ServerStatusValue = iota

	// StatusRunning indicates the server is up and active.
	StatusRunning

	// StatusQuiet indicates the server is up but not active.
	StatusQuiet

	// StatusStopped indicates the server server has been stopped.
	StatusStopped
)

var statuses = []string{
	"idle",
	"running",
	"quiet",
	"stopped",
}

func (s *ServerStatus) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if StatusIdle <= s.val && s.val <= StatusStopped {
		return statuses[s.val]
	}
	return "unknown status"
}

// Get returns the status value.
func (s *ServerStatus) Get() ServerStatusValue {
	s.mu.Lock()
	v := s.val
	s.mu.Unlock()
	return v
}

// Set sets the status value.
func (s *ServerStatus) Set(v ServerStatusValue) {
	s.mu.Lock()
	s.val = v
	s.mu.Unlock()
}
