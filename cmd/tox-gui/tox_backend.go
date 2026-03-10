package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/opd-ai/toxcore"
)

// bootstrapNodes lists well-known public Tox DHT bootstrap nodes.
// These are used on startup to connect the local node to the Tox network.
var bootstrapNodes = []struct {
	host   string
	port   uint16
	pubkey string
}{
	{
		"node.tox.biribiri.org", 33445,
		"F404ABAA1C99A9D37D61AB54898F56793E1DEF8BD46B1038B9D822E8460FAB67",
	},
	{
		"130.133.110.14", 33445,
		"461FA3776EF0FA655F1A05477DF1B3B614F7D6B33955E5D4BE496EFB31C22A11",
	},
	{
		"51.255.69.170", 33445,
		"2B2137E094F743AC8BD44652C55F41DFACC502F125E99E4FE24D40537489E32F",
	},
}

// ToxBackend owns the toxcore.Tox instance and bridges its event callbacks
// to the UI layer.  All public methods are safe to call from any goroutine.
type ToxBackend struct {
	mu    sync.RWMutex
	tox   *toxcore.Tox
	ui    *UI
	stopC chan struct{}
}

// newToxBackend allocates a ToxBackend without starting the Tox network.
func newToxBackend() *ToxBackend {
	return &ToxBackend{
		stopC: make(chan struct{}),
	}
}

// setUI links the UI layer so that callbacks can trigger UI updates.
func (b *ToxBackend) setUI(ui *UI) {
	b.mu.Lock()
	b.ui = ui
	b.mu.Unlock()
}

// ── Initialisation ────────────────────────────────────────────────────────────

// init creates a Tox instance, registers callbacks, and connects to bootstrap
// nodes.  It must be called before run.
func (b *ToxBackend) init() error {
	opts := toxcore.NewOptions()
	opts.UDPEnabled = true
	opts.IPv6Enabled = true

	tox, err := toxcore.New(opts)
	if err != nil {
		return fmt.Errorf("creating Tox instance: %w", err)
	}

	// Store the instance before registering callbacks so that callback
	// closures that call b.tox(...) are safe.
	b.mu.Lock()
	b.tox = tox
	b.mu.Unlock()

	b.registerCallbacks(tox)
	b.bootstrapNetwork(tox)

	// Show our own Tox ID in the footer.
	if b.ui != nil {
		b.ui.SetToxID(tox.SelfGetAddress())
	}

	return nil
}

// registerCallbacks attaches all Tox event handlers.
func (b *ToxBackend) registerCallbacks(tox *toxcore.Tox) {
	// ── Incoming friend request ───────────────────────────────────────────────
	// For simplicity the application auto-accepts every friend request.
	// A production client would present a confirmation dialog instead.
	tox.OnFriendRequest(func(publicKey [32]byte, message string) {
		log.Printf("friend request from %x: %s", publicKey[:8], message)

		friendID, err := tox.AddFriendByPublicKey(publicKey)
		if err != nil {
			log.Printf("accept friend request: %v", err)
			return
		}
		log.Printf("accepted friend, id=%d", friendID)

		if b.ui != nil {
			name := fmt.Sprintf("%x", publicKey[:8])
			b.ui.AddFriendToSidebar(friendID, name, false)
		}
	})

	// ── Incoming message ─────────────────────────────────────────────────────
	tox.OnFriendMessage(func(friendID uint32, message string) {
		if b.ui != nil {
			b.ui.ReceiveMessage(friendID, message)
		}
	})

	// ── Friend connection status ──────────────────────────────────────────────
	tox.OnFriendConnectionStatus(func(friendID uint32, status toxcore.ConnectionStatus) {
		online := status != toxcore.ConnectionNone
		log.Printf("friend %d connection: %v (online=%v)", friendID, status, online)

		if b.ui == nil {
			return
		}

		friends := tox.GetFriends()
		name := fmt.Sprintf("Friend %d", friendID)
		if f, ok := friends[friendID]; ok && f.Name != "" {
			name = f.Name
		}
		b.ui.AddFriendToSidebar(friendID, name, online)
	})

	// ── Self connection status ────────────────────────────────────────────────
	tox.OnConnectionStatus(func(status toxcore.ConnectionStatus) {
		if b.ui == nil {
			return
		}
		var text string
		switch status {
		case toxcore.ConnectionNone:
			text = "Disconnected"
		case toxcore.ConnectionTCP:
			text = "Connected (TCP)"
		case toxcore.ConnectionUDP:
			text = "Connected (UDP)"
		default:
			text = fmt.Sprintf("Unknown status %d", status)
		}
		b.ui.SetStatus(text)
	})
}

// bootstrapNetwork sends DHT bootstrap packets to well-known nodes.
func (b *ToxBackend) bootstrapNetwork(tox *toxcore.Tox) {
	for _, node := range bootstrapNodes {
		if err := tox.Bootstrap(node.host, node.port, node.pubkey); err != nil {
			log.Printf("bootstrap %s: %v", node.host, err)
		}
	}
}

// ── Event loop ────────────────────────────────────────────────────────────────

// run drives the Tox event loop until stop() is called.
// It must be called in its own goroutine.
func (b *ToxBackend) run() {
	b.mu.RLock()
	tox := b.tox
	b.mu.RUnlock()

	if tox == nil {
		return
	}

	for {
		select {
		case <-b.stopC:
			tox.Kill()
			return
		default:
		}

		if !tox.IsRunning() {
			return
		}

		tox.Iterate()
		time.Sleep(tox.IterationInterval())
	}
}

// stop signals the event loop to exit and waits for Tox to shut down.
func (b *ToxBackend) stop() {
	select {
	case <-b.stopC:
		// already closed
	default:
		close(b.stopC)
	}
}

// ── Public API consumed by the UI layer ──────────────────────────────────────

// sendMessage sends a plain-text message to a friend.
func (b *ToxBackend) sendMessage(friendID uint32, message string) error {
	b.mu.RLock()
	tox := b.tox
	b.mu.RUnlock()

	if tox == nil {
		return fmt.Errorf("Tox not initialised")
	}
	return tox.SendFriendMessage(friendID, message)
}

// addFriend sends a friend request to the given Tox address.
func (b *ToxBackend) addFriend(address, message string) error {
	b.mu.RLock()
	tox := b.tox
	b.mu.RUnlock()

	if tox == nil {
		return fmt.Errorf("Tox not initialised")
	}
	_, err := tox.AddFriend(address, message)
	return err
}

// getFriends returns a snapshot of the current friends map.
// Returns nil if Tox is not initialised.
func (b *ToxBackend) getFriends() map[uint32]*toxcore.Friend {
	b.mu.RLock()
	tox := b.tox
	b.mu.RUnlock()

	if tox == nil {
		return nil
	}
	return tox.GetFriends()
}
