package tts_test

import (
	"io"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/Spectral-Kaburu/annoying-sister/internal/tts"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	c := tts.NewClient("/tmp/test.sock")
	require.Equal(t, "/tmp/test.sock", c.SocketPath)
	require.Equal(t, 2*time.Second, c.DialTimeout)
}

func TestClient_Speak_Success(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "tts.sock")

	l, err := net.Listen("unix", sockPath)
	require.NoError(t, err)
	defer l.Close()

	received := make(chan string, 1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf, _ := io.ReadAll(conn)
		received <- string(buf)
	}()

	client := tts.NewClient(sockPath)
	err = client.Speak("Hello world")
	require.NoError(t, err)

	select {
	case msg := <-received:
		require.Equal(t, "speak|Hello world", msg)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for message on unix socket")
	}
}

func TestClient_Speak_MissingSocket(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "nonexistent.sock")
	client := tts.NewClient(sockPath)
	client.DialTimeout = 50 * time.Millisecond

	err := client.Speak("Hello")
	require.Error(t, err)
}
