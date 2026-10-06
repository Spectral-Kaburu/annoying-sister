# Package Reference

This document provides a detailed overview of every package in the `nudged` codebase.

---

## [`cmd/nudged`](file:///home/spectre/Documents/assistant/Annoying-sister/cmd/nudged)

The daemon binary entry point and CLI wizard.

| File | Purpose |
|------|---------|
| [`main.go`](file:///home/spectre/Documents/assistant/Annoying-sister/cmd/nudged/main.go) | Initializes logging, loads config/registry/state, sets up D-Bus, launches the 4 runtime goroutines (`runIdleWatcher`, `runProjectScanner`, `runOnStart`, `runConsumer`), and listens for OS shutdown signals. |
| [`setup.go`](file:///home/spectre/Documents/assistant/Annoying-sister/cmd/nudged/setup.go) | Interactive terminal wizard (`nudged -setup`) powered by Bubble Tea and Lip Gloss. Provides a split-panel interface to browse directories and manage `scan_roots`. |

---

## [`internal/idle`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/idle)

Idle time monitoring, media playback inspection, edge detection, and contextual roasts.

| File | Responsibility |
|------|---------------|
| [`dbus.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/idle/dbus.go) | `DBusReader` queries `org.freedesktop.ScreenSaver.GetSessionIdleTime` over the user session D-Bus. |
| [`decide.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/idle/decide.go) | `Watcher.Decide()` implements pure edge-detection: fires on crossing threshold, repeats while idle if repeat interval has elapsed, resets when user returns. |
| [`mediaplayer.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/idle/mediaplayer.go) | `DBusMediaPlayerReader` inspects `org.mpris.MediaPlayer2.*` bus names on D-Bus to check if media is currently playing (`PlaybackStatus == "Playing"`). |
| [`roast.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/idle/roast.go) | `Roaster` and template pool for generating contextual roast messages when the user is idle while playing media. |
| [`nudge.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/idle/nudge.go) | `NudgeEngine` wraps `Watcher`, `MediaPlayerReader`, and `Roaster`. `Poll(...) (fired bool, roastMsg string)` returns whether to fire and any generated roast. If `fired=true` and `roastMsg=""`, caller renders standard idle templates. |

---

## [`internal/projects`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/projects)

Project discovery, activity computation, and registry management.

| File | Responsibility |
|------|---------------|
| [`store.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/projects/store.go) | Thread-safe `Store` guarding `Registry` (`map[string]Project`), persisted atomically to `projects.json`. Tracks `HasUncommitted` and `UncommittedCount`. |
| [`merge.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/projects/merge.go) | Pure 4-case merge algorithm reconciling discovered paths and git status with existing registry entries. |
| [`discover.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/projects/discover.go) | `Discover()` traverses `scan_roots` with `filepath.WalkDir`, filtering out paths in `ignored_paths`, and locates Git root directories. |
| [`gitstatus.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/projects/gitstatus.go) | `CheckGitStatus()` queries `git status --porcelain` to identify uncommitted changes and uncommitted file count. |
| [`lastactive.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/projects/lastactive.go) | `ComputeLastActive()` walks all files/directories within a project (skipping `.git/`) and identifies the latest mtime. |
| [`dormancy.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/projects/dormancy.go) | `IsDormant()`, `MostDormant()`, and `MostUrgent()` evaluation helpers prioritizing uncommitted dormant projects. |

---

## [`internal/nudge`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/nudge)

Nudge event definitions and template rendering.

| File | Responsibility |
|------|---------------|
| [`event.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/nudge/event.go) | `Event` struct with `Category` (`idle`, `dormant_project`, `on_start`) and `Text`. |
| [`templates.go`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/nudge/templates.go) | Template pools and render functions (`RenderIdle`, `RenderDormant`, `RenderOnStart`) with slot substitution (`{IdleMinutes}`, `{ProjectName}`, `{DormantDays}`). |

---

## [`internal/config`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/config)

Loads and writes `~/.blackboxx/nudge/config.json`. Automatically generates default configuration if missing.

```go
type Config struct {
    ScanRoots                  []string `json:"scan_roots"`
    IgnoredPaths               []string `json:"ignored_paths"`
    IdleThresholdMinutes       int      `json:"idle_threshold_minutes"`
    IdleRepeatIntervalMinutes  int      `json:"idle_repeat_interval_minutes"`
    IdlePollIntervalSeconds    int      `json:"idle_poll_interval_seconds"`
    DormantThresholdDays       int      `json:"dormant_threshold_days"`
    ProjectScanIntervalMinutes int      `json:"project_scan_interval_minutes"`
    SpectreTTSSocketPath       string   `json:"spectretts_socket_path"`
}
```

---

## [`internal/state`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/state)

Persists cooldown state across daemon restarts in `~/.blackboxx/nudge/state.json`.

| Method | Description |
|--------|-------------|
| `IdleLastNudged()` / `SetIdleLastNudged(time.Time)` | Tracks the timestamp of the last idle nudge. |
| `ProjectLastNudgedDate(name)` / `SetProjectLastNudgedDate(name, date)` | Tracks calendar date `"YYYY-MM-DD"` of the last dormant nudge for a given project. |

---

## [`internal/tts`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/tts)

Unix socket client for SpectreTTS.

- Protocol: Sends `speak|<text>` string over a Unix domain socket.
- Per-message connection: Connects, writes payload, and closes per message to avoid stale state if SpectreTTS restarts.

---

## [`internal/paths`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/paths)

Single source of truth for the file system directory layout under `~/.blackboxx/nudge/`.

| Function | Returned Path |
|----------|---------------|
| `NudgeDir()` | `~/.blackboxx/nudge/` |
| `ConfigPath()` | `~/.blackboxx/nudge/config.json` |
| `ProjectsPath()` | `~/.blackboxx/nudge/projects.json` |
| `StatePath()` | `~/.blackboxx/nudge/state.json` |
| `EnsureNudgeDir()` | Creates `~/.blackboxx/nudge/` with `0755` permissions |

---

## [`internal/atomicfile`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/atomicfile)

Atomic file write helper `Write(path, data, perm)` using temporary file creation in the target directory followed by `os.Rename`.

---

## [`internal/logging`](file:///home/spectre/Documents/assistant/Annoying-sister/internal/logging)

Structured logging using Go `log/slog` and `lmittmann/tint`.
- Exposes `Init()` and `For(logger, component)`.
- Standardizes component attribute names across all subsystems.
