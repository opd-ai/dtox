package main

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// AppState holds the shared application state, protected by a read-write mutex
// for safe concurrent access from the UI goroutine and the Tox backend goroutine.
type AppState struct {
	mu              sync.RWMutex
	selfAddress     string
	selfName        string
	selfStatus      string
	connectionState string
	friends         []FriendEntry
	selectedFriend  int
	messages        map[uint32][]ChatMessage
	pendingRequests []PendingRequest
}

// FriendEntry represents a friend in the contact list.
type FriendEntry struct {
	ID               uint32
	Name             string
	StatusMessage    string
	Online           bool
	ConnectionType   string
	PublicKeyShortHex string // First 16 hex chars of the public key, for display
	UnreadCount      int
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Sender    string
	Content   string
	Timestamp time.Time
	Outgoing  bool
}

// PendingRequest represents an incoming friend request that hasn't been accepted yet.
type PendingRequest struct {
	PublicKey [32]byte
	Message   string
	Time      time.Time
}

// NewAppState creates a new initialized AppState.
func NewAppState() *AppState {
	return &AppState{
		selfName:        "Tox User",
		selfStatus:      "Using toxcore-go GUI",
		connectionState: "Disconnected",
		selectedFriend:  -1,
		messages:        make(map[uint32][]ChatMessage),
	}
}

// GetSelfAddress returns the user's Tox address.
func (s *AppState) GetSelfAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selfAddress
}

// SetSelfAddress sets the user's Tox address.
func (s *AppState) SetSelfAddress(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selfAddress = addr
}

// GetSelfName returns the user's display name.
func (s *AppState) GetSelfName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selfName
}

// SetSelfName sets the user's display name.
func (s *AppState) SetSelfName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selfName = name
}

// GetSelfStatus returns the user's status message.
func (s *AppState) GetSelfStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selfStatus
}

// SetSelfStatus sets the user's status message.
func (s *AppState) SetSelfStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selfStatus = status
}

// GetConnectionState returns the current connection state string.
func (s *AppState) GetConnectionState() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectionState
}

// SetConnectionState sets the current connection state string.
func (s *AppState) SetConnectionState(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connectionState = state
}

// GetFriends returns a copy of the friend list.
func (s *AppState) GetFriends() []FriendEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]FriendEntry, len(s.friends))
	copy(result, s.friends)
	return result
}

// SetFriends replaces the friend list.
func (s *AppState) SetFriends(friends []FriendEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.friends = friends
}

// UpdateFriend updates a single friend entry by ID. Returns false if not found.
func (s *AppState) UpdateFriend(id uint32, updater func(*FriendEntry)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.friends {
		if s.friends[i].ID == id {
			updater(&s.friends[i])
			return true
		}
	}
	return false
}

// AddFriend appends a new friend entry.
func (s *AppState) AddFriend(entry FriendEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.friends = append(s.friends, entry)
}

// RemoveFriend removes a friend by ID.
func (s *AppState) RemoveFriend(id uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.friends {
		if s.friends[i].ID == id {
			s.friends = append(s.friends[:i], s.friends[i+1:]...)
			return
		}
	}
}

// GetSelectedFriend returns the index of the currently selected friend.
func (s *AppState) GetSelectedFriend() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedFriend
}

// SetSelectedFriend sets the index of the currently selected friend.
func (s *AppState) SetSelectedFriend(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedFriend = idx
}

// GetSelectedFriendID returns the friend ID of the currently selected friend,
// or (0, false) if no friend is selected.
func (s *AppState) GetSelectedFriendID() (uint32, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.selectedFriend < 0 || s.selectedFriend >= len(s.friends) {
		return 0, false
	}
	return s.friends[s.selectedFriend].ID, true
}

// AppendMessage adds a message to the history for a given friend.
func (s *AppState) AppendMessage(friendID uint32, msg ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[friendID] = append(s.messages[friendID], msg)
}

// GetMessages returns a copy of the message history for a given friend.
func (s *AppState) GetMessages(friendID uint32) []ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs := s.messages[friendID]
	result := make([]ChatMessage, len(msgs))
	copy(result, msgs)
	return result
}

// GetPendingRequests returns a copy of the pending friend requests.
func (s *AppState) GetPendingRequests() []PendingRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]PendingRequest, len(s.pendingRequests))
	copy(result, s.pendingRequests)
	return result
}

// AddPendingRequest appends a new pending friend request.
func (s *AppState) AddPendingRequest(req PendingRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingRequests = append(s.pendingRequests, req)
}

// RemovePendingRequest removes a pending request by public key.
func (s *AppState) RemovePendingRequest(publicKey [32]byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.pendingRequests {
		if s.pendingRequests[i].PublicKey == publicKey {
			s.pendingRequests = append(s.pendingRequests[:i], s.pendingRequests[i+1:]...)
			return
		}
	}
}

// IncrementUnread increments the unread count for a friend.
func (s *AppState) IncrementUnread(friendID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.friends {
		if s.friends[i].ID == friendID {
			s.friends[i].UnreadCount++
			return
		}
	}
}

// ClearUnread resets the unread count for a friend.
func (s *AppState) ClearUnread(friendID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.friends {
		if s.friends[i].ID == friendID {
			s.friends[i].UnreadCount = 0
			return
		}
	}
}

// FormatPublicKeyShort returns the first 16 hex characters of a public key for display.
func FormatPublicKeyShort(key [32]byte) string {
	return fmt.Sprintf("%.16s", hex.EncodeToString(key[:]))
}
