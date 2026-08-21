# nudged

A standalone, long-running Go daemon that speaks **idle** and **dormant-project** nudges
through [SpectreTTS](https://github.com/Spectral-Kaburu/SpectreTTS). Part of the Linus toolchain.

---

## Quick start

```sh
# 1. Build
go build -o ~/.local/bin/nudged ./cmd/nudged

# 2. Interactive setup — pick your project directories in a TUI
nudged -setup

# 3. Run
nudged
```

Ctrl-C or `SIGTERM` for graceful shutdown.

---

## Interactive setup (`nudged -setup`)

Instead of hand-editing JSON, run the setup wizard:

```sh
nudged -setup
```

A full-screen TUI opens with two panels:

| Panel (Tab to switch) | What you do |
|-----------------------|-------------|
| **Left — directory browser** | Navigate your filesystem; press `Enter` on any directory to add it to `scan_roots` |
| **Right — selected roots** | See what's queued; use `↑/↓` to move, `d` or `⌫` to remove an entry |

Key bindings:

| Key | Action |
|-----|--------|
| `Tab` | Switch between browser and list panels |
| `Enter` | Add highlighted directory to scan roots |
| `d` / `⌫` | Remove selected entry from the list |
| `s` / `Ctrl+S` | **Save** and exit |
| `q` / `Esc` | Quit without saving |

The wizard reads your current `config.json` on open and writes back atomically on save —
all other settings (thresholds, intervals, socket path) are preserved.

---

## Configuration

`~/.blackboxx/nudge/config.json` — created with defaults on first run.

```json
{
  "scan_roots": [],
  "ignored_paths": [],
  "idle_threshold_minutes": 20,
  "idle_repeat_interval_minutes": 20,
  "idle_poll_interval_seconds": 30,
  "dormant_threshold_days": 7,
  "project_scan_interval_minutes": 15,
  "spectretts_socket_path": "/tmp/spectretts.sock"
}
```

| Field | Default | What it controls |
|-------|---------|-----------------|
| `scan_roots` | `[]` | Directories walked for project discovery — **set these first** (use `-setup` or edit directly) |
| `ignored_paths` | `[]` | Paths to skip during discovery |
| `idle_threshold_minutes` | `20` | Minutes idle before first nudge |
| `idle_repeat_interval_minutes` | `20` | Minutes between repeated nudges while still idle |
| `idle_poll_interval_seconds` | `30` | D-Bus poll frequency |
| `dormant_threshold_days` | `7` | Days without file changes → project is dormant |
| `project_scan_interval_minutes` | `15` | How often to re-walk `scan_roots` |
| `spectretts_socket_path` | `/tmp/spectretts.sock` | SpectreTTS Unix socket |

> **No hot-reload.** Restart the daemon after editing config.json:
> `systemctl --user restart nudged`

---

## Data files

All files live under `~/.blackboxx/nudge/`.

| File | Owner | Purpose |
|------|-------|---------|
| `config.json` | You | Thresholds, paths, intervals |
| `projects.json` | Daemon (auto) + you (manual) | Project registry with `last_active` timestamps |
| `state.json` | Daemon only | Cooldown bookkeeping |

### Adding a project manually to `projects.json`

Stop the daemon first (atomic-replace races):

```sh
systemctl --user stop nudged
$EDITOR ~/.blackboxx/nudge/projects.json
systemctl --user start nudged
```

```json
{
  "my-project": {
    "path": "/home/spectre/Projects/my-project",
    "summary": "Optional description",
    "source": "manual",
    "last_active": "2026-08-01T10:00:00Z"
  }
}
```

The map key is the display name spoken in nudge phrases.
`source: "manual"` entries are never overwritten by auto-discovery.

### Reset cooldowns

```sh
echo '{}' > ~/.blackboxx/nudge/state.json
```

---

## Logging

Controlled by environment variables, not `config.json` (operational concern):

```sh
NUDGE_LOG_LEVEL=debug   # debug | info | warn | error   (default: info)
NUDGE_LOG_FORMAT=color  # text | json | color            (default: color on TTY, text otherwise)
```

| Format | Use when |
|--------|----------|
| `color` | Running interactively — ANSI colors, `HH:MM:SS` timestamps |
| `text` | `journalctl`, log files — plain `key=value` |
| `json` | Log aggregators, `jq` — structured JSON per line |

Every log line carries a `component` attribute for per-subsystem filtering:

```sh
journalctl --user -u nudged -g 'component=idle_watcher'
journalctl --user -u nudged -o cat | jq 'select(.level=="WARN")'
```

---

## Build / test / run

```sh
go build ./...          # build all packages
go test ./...           # unit tests (no D-Bus, no real socket)
go run ./cmd/nudged     # run directly (color logs auto-enabled on TTY)

NUDGE_LOG_LEVEL=debug go run ./cmd/nudged          # verbose
NUDGE_LOG_FORMAT=text  go run ./cmd/nudged 2>nudge.log   # log to file
```

---

## Deploying (systemd --user)

```sh
go build -o ~/.local/bin/nudged ./cmd/nudged

mkdir -p ~/.config/systemd/user
cp systemd/nudged.service ~/.config/systemd/user/

systemctl --user daemon-reload
systemctl --user enable --now nudged

# Check
systemctl --user status nudged
journalctl --user -u nudged -f
```

Add `Environment=NUDGE_LOG_LEVEL=debug` to the `[Service]` section for persistent
debug logging.

---

## Project layout

```
cmd/nudged/
  main.go        four goroutine drivers + -setup flag
  setup.go       interactive scan_roots TUI wizard (bubbletea)
internal/
  paths/         ~/.blackboxx/nudge/ layout (DataRoot const)
  atomicfile/    temp-file-then-rename write helper
  logging/       slog wrapper — color/text/json, env-driven level
  config/        config.json — load / create defaults / Save
  state/         state.json — cooldown bookkeeping, mutex-guarded
  projects/      discovery, last_active, merge, dormancy logic
  idle/          D-Bus idle reader + pure edge-detection (Watcher)
  nudge/         Event type + template pools / rendering
  tts/           SpectreTTS Unix socket client
systemd/
  nudged.service systemd --user unit
docs/
  nudged.md      Full developer & operator reference
```

Decision logic (`Watcher.Decide`, `Merge`, `IsDormant`) is pure — no D-Bus, no
filesystem, no goroutines — so it's fully unit-testable with plain in-memory
arguments. Goroutine wiring in `main.go` is the only place that touches D-Bus
or tickers.

---

## Design decisions

| Decision | Choice | Why |
|----------|--------|-----|
| New project display name | `filepath.Base(path)` | Rename in `merge.go` (`uniqueKey`) to customize |
| Default `scan_roots` | Empty `[]` | Spec example hardcodes a single user's path — wrong for everyone else; use `-setup` |
| SpectreTTS connection | Per-message connect/close | Avoids reconnect logic if SpectreTTS restarts |
| Key collisions | `-2`, `-3` suffix | Two scan_roots with same-named subdirs → `name` and `name-2` |
| Color logging | `lmittmann/tint` | Zero-dep slog.Handler; auto-off when stderr isn't a TTY |
| Setup wizard | `charmbracelet/bubbletea` + `bubbles/filepicker` | Full-screen TUI directory picker, saves atomically |
| No hot-reload | Restart required | Config changes are infrequent; avoids SIGHUP handler complexity |
| Shutdown | `ctx.Done()` only between events | Spec: never interrupt a mid-write to the TTS socket |
