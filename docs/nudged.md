# nudged — Developer & Operator Documentation

> A long-running Go daemon that speaks idle, media procrastination roasts, and dormant-project nudges through **SpectreTTS**.

---

## What it does

`nudged` monitors system activity and project workspaces, speaking through SpectreTTS when triggers occur:

| Trigger | Condition | Cooldown |
|---------|-----------|----------|
| **Idle nudge** | No keyboard/mouse interaction for `idle_threshold_minutes` | Repeats every `idle_repeat_interval_minutes` while idle; resets on user return |
| **Media roast** | User is idle while a media player (MPRIS) is actively playing | Fired during idle poll with contextual roast referencing the player |
| **Dormant project nudge** | A project in `scan_roots` has had no file edits for `dormant_threshold_days` | Once per calendar day per project |
| **On-start nudge** | Fires once after initial idle check and project scan complete | Never repeats |

---

## Documentation Index

- [**Architecture & Data Flow**](file:///home/spectre/Documents/assistant/Annoying-sister/docs/architecture.md)  
  Goroutine model, pipeline diagrams, SpectreTTS socket serialization, and core invariants.

- [**Package Reference**](file:///home/spectre/Documents/assistant/Annoying-sister/docs/packages.md)  
  Detailed breakdown of `cmd/nudged` and all `internal/*` packages (`idle`, `projects`, `nudge`, `config`, `state`, `tts`, `paths`, `atomicfile`, `logging`).

- [**Configuration & Data Files**](file:///home/spectre/Documents/assistant/Annoying-sister/docs/configuration.md)  
  `config.json` options, data stores (`projects.json`, `state.json`), and manual project registration.

- [**Logging**](file:///home/spectre/Documents/assistant/Annoying-sister/docs/logging.md)  
  Environment variables, format modes (color, text, JSON), log levels, and journalctl filtering.

- [**Operations & Deployment**](file:///home/spectre/Documents/assistant/Annoying-sister/docs/operations.md)  
  Setup wizard (`nudged -setup`), build/test commands, and systemd user service setup.

- [**Design Decisions & Trade-offs**](file:///home/spectre/Documents/assistant/Annoying-sister/docs/design.md)  
  Key architectural choices, edge detection strategy, and rationale behind MPRIS integration.

---

## Quick Start

```sh
# Run interactive setup wizard to select scan roots
go run ./cmd/nudged -setup

# Run daemon with live colored logs
NUDGE_LOG_LEVEL=debug go run ./cmd/nudged

# Run all unit tests
go test ./...
```
