package main

import (
	"fmt"
	"log"
	"time"

	toxcore "github.com/opd-ai/toxcore"
	"github.com/opd-ai/wain"
)

// ToxBackend owns the Tox instance and bridges events to the UI.
type ToxBackend struct {
	tox    *toxcore.Tox
	state  *AppState
	app    *wain.App
	uiRefs *UIRefs
	done   chan struct{}
}

// NewToxBackend creates a new backend, initializes the Tox instance,
// and registers all callbacks.
func NewToxBackend(state *AppState, app *wain.App, uiRefs *UIRefs) (*ToxBackend, error) {
	options := toxcore.NewOptions()
	options.UDPEnabled = true

	tox, err := toxcore.New(options)
	if err != nil {
		return nil, fmt.Errorf("toxcore.New: %w", err)
	}

	if err := tox.SelfSetName(state.GetSelfName()); err != nil {
		log.Printf("SelfSetName: %v", err)
	}
	if err := tox.SelfSetStatusMessage(state.GetSelfStatus()); err != nil {
		log.Printf("SelfSetStatusMessage: %v", err)
	}

	state.SetSelfAddress(tox.SelfGetAddress())

	b := &ToxBackend{
		tox:    tox,
		state:  state,
		app:    app,
		uiRefs: uiRefs,
		done:   make(chan struct{}),
	}

	b.registerCallbacks()
	return b, nil
}

// Start bootstraps to the DHT and starts the Tox event loop goroutine.
func (b *ToxBackend) Start() {
	b.doBootstrap()
	b.syncFriendList()

	go func() {
		for b.tox.IsRunning() {
			select {
			case <-b.done:
				return
			default:
				b.tox.Iterate()
				time.Sleep(b.tox.IterationInterval())
			}
		}
	}()
}

// Stop shuts down the Tox instance and stops the event loop.
func (b *ToxBackend) Stop() {
	select {
	case <-b.done:
		// already stopped
	default:
		close(b.done)
	}
	b.tox.Kill()
}

// SendMessage sends a text message to the given friend.
func (b *ToxBackend) SendMessage(friendID uint32, text string) error {
	if err := b.tox.SendFriendMessage(friendID, text); err != nil {
		return err
	}
	msg := ChatMessage{
		Sender:    "self",
		Content:   text,
		Timestamp: time.Now(),
		Outgoing:  true,
	}
	b.state.AppendMessage(friendID, msg)

	b.app.Notify(func() {
		if id, ok := b.state.GetSelectedFriendID(); ok && id == friendID {
			b.refreshChatView()
		}
	})
	return nil
}

// AddFriend sends a friend request to the given Tox address.
func (b *ToxBackend) AddFriend(address string) error {
	if !isValidToxAddress(address) {
		return fmt.Errorf("invalid Tox address: must be 76 hex characters")
	}
	_, err := b.tox.AddFriend(address, "Friend request from tox-gui")
	if err != nil {
		return err
	}
	b.syncFriendList()
	b.app.Notify(func() {
		b.refreshFriendList()
	})
	return nil
}

// AcceptFriendRequest accepts a pending friend request.
func (b *ToxBackend) AcceptFriendRequest(publicKey [32]byte) error {
	_, err := b.tox.AddFriendByPublicKey(publicKey)
	if err != nil {
		return err
	}
	b.state.RemovePendingRequest(publicKey)
	b.syncFriendList()
	b.app.Notify(func() {
		b.refreshFriendList()
	})
	return nil
}

// registerCallbacks sets up all Tox event callbacks. Each callback updates
// the shared state and then uses app.Notify to safely update UI widgets
// on the UI goroutine.
func (b *ToxBackend) registerCallbacks() {
	b.tox.OnConnectionStatus(func(status toxcore.ConnectionStatus) {
		var label string
		switch status {
		case toxcore.ConnectionUDP:
			label = "Connected (UDP)"
		case toxcore.ConnectionTCP:
			label = "Connected (TCP)"
		default:
			label = "Disconnected"
		}
		b.state.SetConnectionState(label)
		b.app.Notify(func() {
			if b.uiRefs != nil && b.uiRefs.statusLabel != nil {
				b.uiRefs.statusLabel.SetText(label)
			}
		})
	})

	b.tox.OnFriendRequest(func(publicKey [32]byte, message string) {
		b.state.AddPendingRequest(PendingRequest{
			PublicKey: publicKey,
			Message:   message,
			Time:      time.Now(),
		})
		b.app.Notify(func() {
			b.refreshFriendList()
		})
	})

	b.tox.OnFriendMessage(func(friendID uint32, message string) {
		msg := ChatMessage{
			Sender:    friendNameOrID(b.state, friendID),
			Content:   message,
			Timestamp: time.Now(),
			Outgoing:  false,
		}
		b.state.AppendMessage(friendID, msg)

		if id, ok := b.state.GetSelectedFriendID(); ok && id == friendID {
			b.app.Notify(func() {
				b.refreshChatView()
			})
		} else {
			b.state.IncrementUnread(friendID)
			b.app.Notify(func() {
				b.refreshFriendList()
			})
		}
	})

	b.tox.OnFriendConnectionStatus(func(friendID uint32, status toxcore.ConnectionStatus) {
		connType := "Offline"
		online := false
		switch status {
		case toxcore.ConnectionUDP:
			connType = "UDP"
			online = true
		case toxcore.ConnectionTCP:
			connType = "TCP"
			online = true
		}
		b.state.UpdateFriend(friendID, func(f *FriendEntry) {
			f.Online = online
			f.ConnectionType = connType
		})
		b.app.Notify(func() {
			b.refreshFriendList()
		})
	})

	b.tox.OnFriendStatus(func(friendID uint32, _ toxcore.FriendStatus) {
		b.app.Notify(func() {
			b.refreshFriendList()
		})
	})

	b.tox.OnFriendName(func(friendID uint32, name string) {
		b.state.UpdateFriend(friendID, func(f *FriendEntry) {
			f.Name = name
		})
		b.app.Notify(func() {
			b.refreshFriendList()
			if id, ok := b.state.GetSelectedFriendID(); ok && id == friendID {
				b.refreshChatView()
			}
		})
	})

	b.tox.OnFriendStatusMessage(func(friendID uint32, statusMessage string) {
		b.state.UpdateFriend(friendID, func(f *FriendEntry) {
			f.StatusMessage = statusMessage
		})
	})
}

// syncFriendList refreshes the AppState friend list from the Tox instance.
func (b *ToxBackend) syncFriendList() {
	toxFriends := b.tox.GetFriends()
	entries := make([]FriendEntry, 0, len(toxFriends))
	for id, f := range toxFriends {
		connType := "Offline"
		online := false
		switch f.ConnectionStatus {
		case toxcore.ConnectionUDP:
			connType = "UDP"
			online = true
		case toxcore.ConnectionTCP:
			connType = "TCP"
			online = true
		}
		name := f.Name
		if name == "" {
			name = FormatPublicKeyShort(f.PublicKey)
		}
		entries = append(entries, FriendEntry{
			ID:             id,
			Name:           name,
			StatusMessage:  f.StatusMessage,
			Online:         online,
			ConnectionType: connType,
			PublicKeyHex:   FormatPublicKeyShort(f.PublicKey),
		})
	}
	b.state.SetFriends(entries)
}

// refreshFriendList rebuilds the sidebar friend list widgets.
func (b *ToxBackend) refreshFriendList() {
	if b.uiRefs == nil || b.uiRefs.friendList == nil {
		return
	}
	rebuildFriendList(b.uiRefs, b.state, b)
}

// refreshChatView rebuilds the chat message widgets.
func (b *ToxBackend) refreshChatView() {
	if b.uiRefs == nil || b.uiRefs.messageList == nil {
		return
	}
	rebuildChatView(b.uiRefs, b.state)
}

// isValidToxAddress validates that a string is a 76-character hex Tox address.
func isValidToxAddress(addr string) bool {
	if len(addr) != 76 {
		return false
	}
	for _, c := range addr {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// friendNameOrID returns the display name for a friend, falling back to the ID.
func friendNameOrID(state *AppState, friendID uint32) string {
	friends := state.GetFriends()
	for _, f := range friends {
		if f.ID == friendID {
			if f.Name != "" {
				return f.Name
			}
			return f.PublicKeyHex
		}
	}
	return fmt.Sprintf("Friend#%d", friendID)
}
