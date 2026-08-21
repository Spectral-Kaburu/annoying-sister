package idle

import (
	"strings"

	"github.com/godbus/dbus/v5"
)

// MediaPlayerReader reports whether any MPRIS-compliant media player is
// currently playing. It's an interface so NudgeEngine can be unit-tested
// with a fake instead of a real D-Bus connection, mirroring the Reader
// interface in dbus.go.
type MediaPlayerReader interface {
	// NowPlaying returns (true, playerName, nil) if at least one MPRIS
	// player reports PlaybackStatus == "Playing". playerName is a
	// human-friendly guess derived from the bus name (e.g. "Vlc",
	// "Mpv") for use in nudge text. If nothing is playing, returns
	// (false, "", nil). A player that's open but paused does not count.
	NowPlaying() (bool, string, error)
}

// DBusMediaPlayerReader enumerates org.mpris.MediaPlayer2.* bus names on
// the session bus and checks each one's PlaybackStatus property. Every
// MPRIS-compliant player (mpv, VLC, Celluloid, Firefox/Chrome <video>,
// Spotify, ...) registers a bus name under this prefix, so this is a
// generic "is anything playing" check with no per-app special-casing.
type DBusMediaPlayerReader struct {
	conn *dbus.Conn
}

// NewDBusMediaPlayerReader wraps an existing session bus connection.
// Reuses the same *dbus.Conn as DBusReader rather than opening a second
// connection — one long-lived session bus connection per process is
// enough for both idle and media-player polling.
func NewDBusMediaPlayerReader(conn *dbus.Conn) *DBusMediaPlayerReader {
	return &DBusMediaPlayerReader{conn: conn}
}

const mprisPrefix = "org.mpris.MediaPlayer2."

// NowPlaying scans all MPRIS bus names for one reporting PlaybackStatus
// == "Playing". Returns on the first match; if you're running two players
// at once and want to know about both, this isn't the tool for that.
func (r *DBusMediaPlayerReader) NowPlaying() (bool, string, error) {
	if r == nil || r.conn == nil {
		return false, "", nil
	}
	var names []string
	if err := r.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names); err != nil {
		return false, "", err
	}

	for _, name := range names {
		if !strings.HasPrefix(name, mprisPrefix) {
			continue
		}

		obj := r.conn.Object(name, "/org/mpris/MediaPlayer2")
		variant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
		if err != nil {
			// Player may have just dropped its bus name, or doesn't
			// implement the Player interface yet mid-startup. Skip
			// rather than fail the whole scan over one flaky player.
			continue
		}

		status, ok := variant.Value().(string)
		if !ok || status != "Playing" {
			continue
		}

		return true, prettyPlayerName(name), nil
	}

	return false, "", nil
}

// prettyPlayerName turns "org.mpris.MediaPlayer2.vlc.instance1234" into
// "Vlc" — good enough for nudge text without a hardcoded per-app lookup
// table. Bus names sometimes carry an instance suffix (e.g. Chrome tabs
// each get their own ".instanceN"); only the first segment is the player
// identity.
func prettyPlayerName(busName string) string {
	rest := strings.TrimPrefix(busName, mprisPrefix)
	if idx := strings.Index(rest, "."); idx != -1 {
		rest = rest[:idx]
	}
	if rest == "" {
		return "something"
	}
	return strings.ToUpper(rest[:1]) + rest[1:]
}
