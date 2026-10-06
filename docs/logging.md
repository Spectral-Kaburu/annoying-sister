# Logging

`nudged` uses Go standard library structured logging (`log/slog`) with ANSI color support via `lmittmann/tint`.

---

## Configuration via Environment Variables

Logging is configured via process environment variables rather than `config.json`:

```sh
NUDGE_LOG_LEVEL=debug   # debug | info | warn | error (default: info)
NUDGE_LOG_FORMAT=color  # color | text | json (default: color on TTY, text otherwise)
```

---

## Log Formats

| Format | Output Characteristics | Typical Use Case |
|--------|------------------------|------------------|
| `color` | Human-readable terminal output with ANSI colors and `HH:MM:SS` timestamps. | Direct interactive execution in a terminal. |
| `text` | Key-value structured text (`level=INFO msg=... component=...`). | `systemd` journal and plain log files. |
| `json` | Structured JSON per line. | Ingestion into log collectors or parsing with `jq`. |

> [!NOTE]
> Auto-detection: When `NUDGE_LOG_FORMAT` is unset, `color` is selected automatically if stderr is attached to a terminal (TTY); otherwise, it falls back to `text`.

---

## Log Levels

| Level | Description |
|-------|-------------|
| `debug` | Idle poll ticks, cache queries, individual decision results, scan timing details. |
| `info` | Daemon lifecycle events (start/stop), fired nudges, scan summaries, configuration loads. |
| `warn` | Non-fatal issues (SpectreTTS socket unreachable, D-Bus poll glitch, single project mtime failure). |
| `error` | Fatal conditions (D-Bus connection failure at boot, corrupt unreadable JSON stores). |

---

## Component Tags

Every log record includes a `component` attribute:

```
component = main | config | state | projects | idle_watcher | project_scanner | on_start | tts_consumer
```

### Filtering in Journalctl

Filter logs by component using journalctl grep:

```sh
# Follow idle watcher activity live
journalctl --user -u nudged -g 'component=idle_watcher' -f

# Inspect warnings across all components
journalctl --user -u nudged -g 'level=WARN'
```
