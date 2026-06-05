package bridge

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/opd-ai/toxcore/transport"
)

// SOCKS5Handler processes incoming SOCKS5 connections and routes them through
// the appropriate transport (Tox friends or direct Tor).
type SOCKS5Handler struct {
	client    net.Conn
	transport *transport.MultiTransport
	failover  *FailoverState
}

// SOCKS5 protocol constants
const (
	// SOCKS5 version
	SOCKS5Version = 0x05
	// SOCKS5 command opcodes
	SOCKS5CmdConnect  = 0x01
	SOCKS5CmdBind     = 0x02
	SOCKS5CmdUDPAssoc = 0x03
	// SOCKS5 address types
	SOCKS5AddrTypeIPv4   = 0x01
	SOCKS5AddrTypeDomain = 0x03
	SOCKS5AddrTypeIPv6   = 0x04
	// SOCKS5 reply codes
	SOCKS5ReplySuccess              = 0x00
	SOCKS5ReplyServerFailure        = 0x01
	SOCKS5ReplyConnectionNotAllowed = 0x02
	SOCKS5ReplyNetworkUnreachable   = 0x03
	SOCKS5ReplyHostUnreachable      = 0x04
	SOCKS5ReplyConnectionRefused    = 0x05
	SOCKS5ReplyTTLExpired           = 0x06
	SOCKS5ReplyCmdNotSupported      = 0x07
	SOCKS5ReplyAddrTypeNotSupported = 0x08
	// Authentication methods
	SOCKS5AuthNoAuth = 0x00
	SOCKS5AuthNone   = 0xFF // No acceptable methods
)

// NewSOCKS5Handler creates a new SOCKS5 connection handler
func NewSOCKS5Handler(client net.Conn, tm *transport.MultiTransport, fs *FailoverState) *SOCKS5Handler {
	return &SOCKS5Handler{
		client:    client,
		transport: tm,
		failover:  fs,
	}
}

// Handle processes a complete SOCKS5 connection session
func (h *SOCKS5Handler) Handle() error {
	// Step 1: Read and respond to SOCKS5 greeting
	if err := h.handleGreeting(); err != nil {
		return fmt.Errorf("greeting failed: %w", err)
	}

	// Step 2: Read and process SOCKS5 request
	cmd, addr, err := h.readRequest()
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	// Step 3: Process command based on failover state
	switch cmd {
	case SOCKS5CmdConnect:
		return h.handleConnect(addr)
	case SOCKS5CmdBind:
		// BIND not supported in this bridge
		if err := h.sendReply(byte(SOCKS5ReplyCmdNotSupported), "", 0); err != nil {
			log.Printf("[SOCKS5] Failed to send unsupported command reply: %v", err)
		}
		return fmt.Errorf("BIND command not supported")
	case SOCKS5CmdUDPAssoc:
		// UDP_ASSOCIATE not supported in this bridge
		if err := h.sendReply(byte(SOCKS5ReplyCmdNotSupported), "", 0); err != nil {
			log.Printf("[SOCKS5] Failed to send unsupported command reply: %v", err)
		}
		return fmt.Errorf("UDP_ASSOCIATE command not supported")
	default:
		if err := h.sendReply(byte(SOCKS5ReplyCmdNotSupported), "", 0); err != nil {
			log.Printf("[SOCKS5] Failed to send invalid command reply: %v", err)
		}
		return fmt.Errorf("unknown command: %d", cmd)
	}
}

// handleGreeting processes the SOCKS5 greeting (authentication method negotiation)
func (h *SOCKS5Handler) handleGreeting() error {
	// Read greeting header: [version, nmethods]
	header := make([]byte, 2)
	if _, err := io.ReadFull(h.client, header); err != nil {
		return fmt.Errorf("failed to read greeting header: %w", err)
	}

	if header[0] != SOCKS5Version {
		return fmt.Errorf("invalid SOCKS5 version: %d", header[0])
	}

	nmethods := int(header[1])
	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(h.client, methods); err != nil {
		return fmt.Errorf("failed to read greeting methods: %w", err)
	}

	selected := byte(SOCKS5AuthNone)
	for _, m := range methods {
		if m == SOCKS5AuthNoAuth {
			selected = SOCKS5AuthNoAuth
			break
		}
	}

	if _, err := h.client.Write([]byte{SOCKS5Version, selected}); err != nil {
		return fmt.Errorf("failed to send greeting response: %w", err)
	}
	if selected == SOCKS5AuthNone {
		return fmt.Errorf("no acceptable auth methods")
	}

	return nil
}

// readRequest reads and parses a SOCKS5 request
// Returns: command, destination address (host:port), error
func (h *SOCKS5Handler) readRequest() (byte, string, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(h.client, buf); err != nil {
		return 0, "", fmt.Errorf("failed to read request header: %w", err)
	}

	// [version, command, reserved, address_type]
	if buf[0] != SOCKS5Version {
		return 0, "", fmt.Errorf("invalid version in request: %d", buf[0])
	}

	cmd := buf[1]
	// buf[2] is reserved
	addrType := buf[3]

	// Parse destination address based on type
	var addr string
	var port uint16

	switch addrType {
	case SOCKS5AddrTypeIPv4:
		// IPv4: 4 bytes for address, 2 bytes for port
		addrBuf := make([]byte, 6)
		if _, err := io.ReadFull(h.client, addrBuf); err != nil {
			return 0, "", fmt.Errorf("failed to read IPv4 address: %w", err)
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", addrBuf[0], addrBuf[1], addrBuf[2], addrBuf[3])
		port = binary.BigEndian.Uint16(addrBuf[4:6])

	case SOCKS5AddrTypeDomain:
		// Domain: 1 byte length, domain string, 2 bytes port
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(h.client, lenBuf); err != nil {
			return 0, "", fmt.Errorf("failed to read domain length: %w", err)
		}
		domainLen := int(lenBuf[0])

		domainBuf := make([]byte, domainLen+2)
		if _, err := io.ReadFull(h.client, domainBuf); err != nil {
			return 0, "", fmt.Errorf("failed to read domain: %w", err)
		}
		addr = string(domainBuf[:domainLen])
		port = binary.BigEndian.Uint16(domainBuf[domainLen:])

	case SOCKS5AddrTypeIPv6:
		// IPv6: 16 bytes for address, 2 bytes for port
		addrBuf := make([]byte, 18)
		if _, err := io.ReadFull(h.client, addrBuf); err != nil {
			return 0, "", fmt.Errorf("failed to read IPv6 address: %w", err)
		}
		// Format IPv6 address
		addr = net.IP(addrBuf[:16]).String()
		port = binary.BigEndian.Uint16(addrBuf[16:18])

	default:
		if err := h.sendReply(byte(SOCKS5ReplyAddrTypeNotSupported), "", 0); err != nil {
			log.Printf("[SOCKS5] Failed to send unsupported addr type reply: %v", err)
		}
		return 0, "", fmt.Errorf("unsupported address type: %d", addrType)
	}

	return cmd, net.JoinHostPort(addr, fmt.Sprintf("%d", port)), nil
}

// handleConnect processes a SOCKS5 CONNECT command
// Routes the connection through the failover state machine
func (h *SOCKS5Handler) handleConnect(targetAddr string) error {
	// Try to connect based on failover state
	var targetConn net.Conn
	var err error

	// Use failover routing: prefer Tox friends, fall back to direct
	if h.failover.ShouldUseToxFriends() {
		// Try to connect through Tox friends first
		// For now, we route through the multi-transport (future: add friend-specific routing)
		log.Printf("[Bridge] Routing through Tox friends: %s", targetAddr)
		targetConn, err = h.transport.Dial(targetAddr)
	} else {
		// Direct routing (Tor or IP)
		log.Printf("[Bridge] Direct routing: %s", targetAddr)
		targetConn, err = h.transport.Dial(targetAddr)
	}

	// Handle connection errors
	if err != nil {
		log.Printf("[Bridge] Connection failed to %s: %v", targetAddr, err)
		reply := byte(SOCKS5ReplyConnectionRefused)
		if err.Error() == "network unreachable" {
			reply = byte(SOCKS5ReplyNetworkUnreachable)
		}
		if err := h.sendReply(reply, "", 0); err != nil {
			log.Printf("[SOCKS5] Failed to send connection error reply: %v", err)
		}
		return err
	}
	defer targetConn.Close()

	// Send success reply
	if err := h.sendReply(byte(SOCKS5ReplySuccess), "", 0); err != nil {
		return fmt.Errorf("failed to send connect reply: %w", err)
	}

	// Relay data between client and target
	// This is a bidirectional relay that continues until the connection closes
	if err := h.relay(h.client, targetConn); err != nil {
		log.Printf("[Bridge] Relay error: %v", err)
	}

	return nil
}

// sendReply sends a SOCKS5 reply to the client
// For simplicity, we always respond with IPv4 0.0.0.0:0 as the bind address
func (h *SOCKS5Handler) sendReply(reply byte, addr string, port uint16) error {
	// [version, reply, reserved, address_type, address, port]
	buf := []byte{
		SOCKS5Version,
		reply,
		0x00, // reserved
		SOCKS5AddrTypeIPv4,
		0, 0, 0, 0, // IPv4 address (0.0.0.0)
		0, 0, // port (0)
	}

	if _, err := h.client.Write(buf); err != nil {
		return fmt.Errorf("failed to write reply: %w", err)
	}

	return nil
}

// relay copies data bidirectionally between client and target connections
func (h *SOCKS5Handler) relay(client, target net.Conn) error {
	errChan := make(chan error, 2)

	// Copy from client to target
	go func() {
		_, copyErr := io.Copy(target, client)
		if tcp, ok := target.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		}
		if copyErr != nil && copyErr != io.EOF {
			errChan <- copyErr
			return
		}
		errChan <- nil
	}()

	// Copy from target to client
	go func() {
		_, copyErr := io.Copy(client, target)
		if tcp, ok := client.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		}
		if copyErr != nil && copyErr != io.EOF {
			errChan <- copyErr
			return
		}
		errChan <- nil
	}()

	// Wait for both goroutines to finish
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}
