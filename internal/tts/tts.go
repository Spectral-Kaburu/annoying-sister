// Package tts writes rendered nudge text to the SpectreTTS Unix socket
// using its fire-and-forget "speak|<text>" protocol.
package tts

import (
	"fmt"
	"net"
	"time"
)

// Client writes messages to a SpectreTTS Unix socket, one at a time.
// There is no established connection-reuse pattern to match elsewhere in
// the codebase yet (this is a new standalone project), so Client opens a
// fresh connection per message and closes it — the simplest option the
// spec calls out as acceptable, and it sidesteps having to detect/recover
// a stale persistent connection if SpectreTTS restarts.
type Client struct {
	SocketPath string
	// DialTimeout bounds how long a single Speak call may block trying to
	// connect, so a wedged/absent SpectreTTS can't stall the consumer
	// goroutine indefinitely.
	DialTimeout time.Duration
}

// NewClient returns a Client for the given socket path with a sensible
// default dial timeout.
func NewClient(socketPath string) *Client {
	return &Client{SocketPath: socketPath, DialTimeout: 2 * time.Second}
}

// Speak writes "speak|<text>" to the socket and closes the connection.
// Errors (SpectreTTS not running, socket missing, etc.) are returned to
// the caller to log; per the spec this is fire-and-forget at the protocol
// level, but the consumer goroutine still wants to know a write failed so
// it can log and move on to the next queued event rather than silently
// dropping nudges.
func (c *Client) Speak(text string) error {
	conn, err := net.DialTimeout("unix", c.SocketPath, c.DialTimeout)
	if err != nil {
		return fmt.Errorf("dial spectretts socket %q: %w", c.SocketPath, err)
	}
	defer conn.Close()

	msg := "speak|" + text
	if _, err := conn.Write([]byte(msg)); err != nil {
		return fmt.Errorf("write to spectretts socket %q: %w", c.SocketPath, err)
	}
	return nil
}
