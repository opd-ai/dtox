package bridge

import (
	"log"
	"sync"
	"time"
)

// RouteState represents the current routing decision in the failover state machine.
// This determines how connections are routed:
// - RouteToxFriends: Attempt to route through available Tox friend bridges
// - RouteDirect: Fall back to direct Tor or IP routing
type RouteState int

const (
	// RouteToxFriends means try to route through active Tox friend bridges first
	RouteToxFriends RouteState = iota
	// RouteDirect means route directly through Tor or IP transports (failover mode)
	RouteDirect
)

// FailoverState implements the failover state machine for automatic bridge routing.
//
// State Machine:
// 1. ToxFriends Available → RouteToxFriends
//    - Try to route connections through available Tox friend bridges
//    - If friends go offline, transition to RouteDirect
//
// 2. No ToxFriends → RouteDirect
//    - Route directly through Tor or IP transports
//    - If Tox friends come online, transition back to RouteToxFriends
//
// 3. Tor Unavailable → Check Periodically
//    - Log warnings and continue attempting Tor until available
//
// The state machine is thread-safe and updates are idempotent.
type FailoverState struct {
	mu              sync.RWMutex
	currentState    RouteState
	activeFriends   int
	torAvailable    bool
	lastUpdate      time.Time
}

// NewFailoverState creates a new failover state machine, initialized to direct routing mode
func NewFailoverState() *FailoverState {
	return &FailoverState{
		currentState:  RouteDirect, // Start in direct mode until Tox friends appear
		activeFriends: 0,
		torAvailable:  true,
		lastUpdate:    time.Now(),
	}
}

// Update processes a state transition event in the failover state machine.
// This should be called periodically with the current number of active Tox friends
// and the availability of Tor. This implements automatic failover logic:
//
// Transition Rules:
// - If activeFriends > 0 and currentState != RouteToxFriends: transition to RouteToxFriends
// - If activeFriends == 0 and currentState != RouteDirect: transition to RouteDirect
// - Update Tor availability for fallback decisions
//
// Parameters:
//   - activeFriends: Number of currently available Tox friend bridges (0 to N)
//   - torAvailable: Whether Tor is currently available for fallback
func (fs *FailoverState) Update(activeFriends int, torAvailable bool) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	oldState := fs.currentState
	fs.activeFriends = activeFriends
	fs.torAvailable = torAvailable

	// State machine transitions
	if activeFriends > 0 {
		// Tox friends available: prefer them as primary route
		fs.currentState = RouteToxFriends
	} else {
		// No Tox friends: fall back to direct routing
		fs.currentState = RouteDirect
	}

	// Log state transitions for visibility
	if oldState != fs.currentState {
		stateStr := func(s RouteState) string {
			if s == RouteToxFriends {
				return "RouteToxFriends"
			}
			return "RouteDirect"
		}
		log.Printf("[Bridge] Failover state transition: %s → %s (friends=%d, tor=%v)",
			stateStr(oldState), stateStr(fs.currentState), activeFriends, torAvailable)
	}

	fs.lastUpdate = time.Now()
}

// GetCurrentState returns the current routing state
func (fs *FailoverState) GetCurrentState() RouteState {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.currentState
}

// GetActiveFriendCount returns the number of active Tox friend bridges
func (fs *FailoverState) GetActiveFriendCount() int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.activeFriends
}

// IsTorAvailable returns whether Tor is currently available
func (fs *FailoverState) IsTorAvailable() bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.torAvailable
}

// GetLastUpdate returns the timestamp of the last state machine update
func (fs *FailoverState) GetLastUpdate() time.Time {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.lastUpdate
}

// ShouldUseToxFriends returns true if the state machine currently prefers Tox friend routes
func (fs *FailoverState) ShouldUseToxFriends() bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.currentState == RouteToxFriends && fs.activeFriends > 0
}

// ShouldUseDirect returns true if the state machine has transitioned to direct routing
// (either no Tox friends or explicit fallback mode)
func (fs *FailoverState) ShouldUseDirect() bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.currentState == RouteDirect
}
