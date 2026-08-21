// Package logging sets up Nudge's structured logger. All components log
// through *slog.Logger with a "component" attribute rather than
// unstructured log.Printf calls, so log output is filterable and
// machine-parseable (e.g. by journald/jq) without grepping message text.
//
// Level and format are controlled by environment variables rather than
// config.json: they're an operational concern (how do I debug this right
// now), not one of the spec's user-adjustable thresholds/intervals/paths.
//
//	NUDGE_LOG_LEVEL  debug | info | warn | error   (default: info)
//	NUDGE_LOG_FORMAT text | json | color            (default: color when stderr
//	                                                  is a TTY, text otherwise)
//
// text is meant for interactive use and `journalctl` output; json is for
// piping into a log aggregator or jq; color gives human-readable,
// ANSI-colored terminal output when running interactively.
package logging

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"golang.org/x/sys/unix"
)

// Init builds the root logger from environment variables, sets it as
// slog's package-level default (so any incidental slog.Info/etc. call
// anywhere in the program is still captured), and returns it. Call once
// at the top of main().
func Init() *slog.Logger {
	level := parseLevel(os.Getenv("NUDGE_LOG_LEVEL"))
	format := strings.ToLower(strings.TrimSpace(os.Getenv("NUDGE_LOG_FORMAT")))

	// Auto-enable color when stderr is a real TTY and no explicit format was set.
	if format == "" {
		if isTTY(os.Stderr) {
			format = "color"
		} else {
			format = "text"
		}
	}

	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	case "color":
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.TimeOnly,
		})
	default: // "text" or anything unrecognized
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// isTTY reports whether f is connected to a real terminal.
func isTTY(f *os.File) bool {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	return err == nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "", "info":
		return slog.LevelInfo
	default:
		// Unrecognized value: fall back to info rather than erroring out
		// over a logging-config typo.
		return slog.LevelInfo
	}
}

// Component-name constants, used both for the "component" attribute and
// kept here so call sites can't typo a component name inconsistently
// across log lines.
const (
	ComponentMain     = "main"
	ComponentConfig   = "config"
	ComponentState    = "state"
	ComponentProjects = "projects"
	ComponentIdle     = "idle_watcher"
	ComponentScanner  = "project_scanner"
	ComponentOnStart  = "on_start"
	ComponentConsumer = "tts_consumer"
)

// For is a small convenience wrapper so call sites read
// logging.For(logger, logging.ComponentIdle) instead of repeating
// logger.With("component", ...) everywhere.
func For(base *slog.Logger, component string) *slog.Logger {
	return base.With("component", component)
}
