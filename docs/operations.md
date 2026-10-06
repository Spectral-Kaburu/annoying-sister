# Operations & Deployment

`nudged` is designed as a headless background service with an interactive setup wizard and standard systemd integration.

---

## Interacting with the Daemon

### 1. Interactive Setup Wizard

```sh
nudged -setup
```

A full-screen terminal UI for managing `scan_roots`:
- `Tab`: Switch between the left Directory Picker and right Selected Roots panels.
- `Enter`: Add current selected directory to `scan_roots`.
- `d` / `Backspace`: Remove selected root (in right panel).
- `s` / `Ctrl+S`: Save configuration and exit.
- `q` / `Esc`: Exit without saving.

---

### 2. Manual Configuration & Project Management

```sh
# Edit configuration
$EDITOR ~/.blackboxx/nudge/config.json
systemctl --user restart nudged

# Manually edit project descriptions or entries
systemctl --user stop nudged
$EDITOR ~/.blackboxx/nudge/projects.json
systemctl --user start nudged
```

---

### 3. Monitoring & Logs

```sh
# Follow live colored logs
NUDGE_LOG_FORMAT=color journalctl --user -u nudged -f

# Filter error logs with jq
journalctl --user -u nudged -o cat | jq 'select(.level == "ERROR")'

# Run interactively with debug output
NUDGE_LOG_LEVEL=debug NUDGE_LOG_FORMAT=color go run ./cmd/nudged
```

---

### 4. Cooldown Reset & Testing

```sh
# Reset cooldowns
echo '{}' > ~/.blackboxx/nudge/state.json

# Test SpectreTTS socket directly
echo -n 'speak|Hello from nudge test' | nc -U /tmp/spectretts.sock
```

---

## Build, Test & Run

```sh
# Build all packages and binaries
go build ./...

# Run the test suite
go test ./...

# Run binary with debug logs
NUDGE_LOG_LEVEL=debug go run ./cmd/nudged
```

---

## Deployment with systemd

### 1. Build and install binary

```sh
mkdir -p ~/.local/bin
go build -o ~/.local/bin/nudged ./cmd/nudged
```

### 2. Install user service unit

```sh
mkdir -p ~/.config/systemd/user
cp systemd/nudged.service ~/.config/systemd/user/
```

### 3. Enable and start service

```sh
systemctl --user daemon-reload
systemctl --user enable --now nudged
```

### 4. Verify service status

```sh
systemctl --user status nudged
journalctl --user -u nudged -f
```
