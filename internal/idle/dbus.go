package idle

import (
	"time"

	"github.com/godbus/dbus/v5"
)

// Reader reports the user's current system idle time. It's an interface
// so the decision logic (Watcher, in decide.go) can be unit-tested with a
// fake implementation instead of a real D-Bus connection.
type Reader interface {
	GetIdleTime() (time.Duration, error)
}

// DBusReader queries org.gnome.Mutter.IdleMonitor.GetIdletime over the
// session bus. This is a GNOME/Mutter-specific interface, not a
// cross-compositor Wayland standard — acceptable per the spec since
// GNOME/Mutter is the target environment.
//
// A single long-lived connection is opened at construction and reused for
// every poll, which is the right choice at the spec's 30s default poll
// interval (opening a fresh session bus connection every 30s would be
// wasteful and adds needless failure surface).
type DBusReader struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

// NewDBusReader opens a private session bus connection and binds to the
// IdleMonitor object. The caller should call Close when done (typically at
// daemon shutdown).
func NewDBusReader() (*DBusReader, error) {
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil, err
	}
	if err := conn.Auth(nil); err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.Hello(); err != nil {
		conn.Close()
		return nil, err
	}

	obj := conn.Object("org.gnome.Mutter.IdleMonitor", "/org/gnome/Mutter/IdleMonitor/Core")
	return &DBusReader{conn: conn, obj: obj}, nil
}

// GetIdleTime returns the current idle duration as reported by Mutter.
func (r *DBusReader) GetIdleTime() (time.Duration, error) {
	var idleMillis uint64
	if err := r.obj.Call("org.gnome.Mutter.IdleMonitor.GetIdletime", 0).Store(&idleMillis); err != nil {
		return 0, err
	}
	return time.Duration(idleMillis) * time.Millisecond, nil
}

// Close closes the underlying D-Bus connection.
func (r *DBusReader) Close() error {
	return r.conn.Close()
}

// Conn returns the underlying D-Bus connection.
func (r *DBusReader) Conn() *dbus.Conn {
	if r == nil {
		return nil
	}
	return r.conn
}

