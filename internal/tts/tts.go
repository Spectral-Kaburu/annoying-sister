// Package tts writes rendered nudge text to the SpectreTTS Unix socket
// using its fire-and-forget "speak|<text>" protocol.
package tts

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CandidateSocketPaths returns standard places where SpectreTTS might create its socket.
func CandidateSocketPaths() []string {
	paths := []string{
		"/tmp/spectretts.sock",
		"/tmp/spectre.sock",
		"/tmp/tts.sock",
	}
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		paths = append(paths,
			filepath.Join(runtimeDir, "spectretts.sock"),
			filepath.Join(runtimeDir, "spectretts/spectretts.sock"),
			filepath.Join(runtimeDir, "spectre.sock"),
			filepath.Join(runtimeDir, "tts.sock"),
		)
	}
	if uid := os.Getuid(); uid >= 0 {
		paths = append(paths,
			fmt.Sprintf("/run/user/%d/spectretts.sock", uid),
			fmt.Sprintf("/run/user/%d/spectretts/spectretts.sock", uid),
			fmt.Sprintf("/run/user/%d/spectre.sock", uid),
			fmt.Sprintf("/run/user/%d/tts.sock", uid),
		)
	}
	return paths
}

// DiscoverSocket attempts to find an existing active SpectreTTS Unix socket.
// If preferPath is explicitly provided and not default, it is respected directly.
func DiscoverSocket(preferPath string) string {
	if preferPath != "" && preferPath != "/tmp/spectretts.sock" {
		return preferPath
	}
	if preferPath == "/tmp/spectretts.sock" {
		if fi, err := os.Stat(preferPath); err == nil && (fi.Mode()&os.ModeSocket != 0 || fi.Mode().IsRegular()) {
			return preferPath
		}
	}
	for _, p := range CandidateSocketPaths() {
		if fi, err := os.Stat(p); err == nil && (fi.Mode()&os.ModeSocket != 0 || fi.Mode().IsRegular()) {
			return p
		}
	}
	if preferPath != "" {
		return preferPath
	}
	return "/tmp/spectretts.sock"
}

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

// resolveSocket returns the configured socket path if it exists, or searches known candidates.
func (c *Client) resolveSocket() string {
	return DiscoverSocket(c.SocketPath)
}

// Speak writes "speak|<text>" to the socket and closes the connection.
// Errors (SpectreTTS not running, socket missing, etc.) are returned to
// the caller to log; per the spec this is fire-and-forget at the protocol
// level, but the consumer goroutine still wants to know a write failed so
// it can log and move on to the next queued event rather than silently
// dropping nudges.
func (c *Client) Speak(text string) error {
	sockPath := c.resolveSocket()
	conn, err := net.DialTimeout("unix", sockPath, c.DialTimeout)
	if err != nil {
		return fmt.Errorf("dial spectretts socket %q: %w", sockPath, err)
	}
	defer conn.Close()

	msg := "speak|" + text
	if _, err := conn.Write([]byte(msg)); err != nil {
		return fmt.Errorf("write to spectretts socket %q: %w", sockPath, err)
	}
	return nil
}
