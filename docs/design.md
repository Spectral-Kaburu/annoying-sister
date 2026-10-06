# Design Decisions & Trade-offs

This document outlines key technical decisions and architectural trade-offs in `nudged`.

---

## Decisions & Trade-offs

| Area | Decision | Rationale |
|------|----------|-----------|
| **Project display name** | `filepath.Base(path)` | Derived directly from project folder name. Easily overridden by modifying the key name in `projects.json`. |
| **Default `scan_roots`** | Empty `[]` | Prevents scanning arbitrary home directories on initial start without user configuration. |
| **SpectreTTS socket connection** | Per-message connect and close | Eliminates complex reconnect and health-check logic if SpectreTTS restarts. Socket dial timeout prevents blocking. |
| **Key name collisions** | `-2`, `-3` numeric suffixes | Avoids overwriting if two different `scan_roots` share identical subfolder names. |
| **Configuration reloading** | Service restart required | Eliminates `SIGHUP` and runtime race condition complexity; configuration changes are infrequent. |
| **Graceful shutdown** | Check `ctx.Done()` between events | Ensures no mid-write interruption to disk or the TTS socket, avoiding corrupted state. |
| **Color logging** | `lmittmann/tint` | Zero-dependency drop-in for `log/slog` with automatic TTY detection. |
| **MPRIS media detection** | `DBusMediaPlayerReader` reuses existing D-Bus connection | Queries all `org.mpris.MediaPlayer2.*` buses using the same shared session connection rather than establishing a second D-Bus connection. |
| **Roast vs. idle template dispatch** | `NudgeEngine.Poll` returns `""` when no media active | Keeps `NudgeEngine` independent of `nudge.RenderIdle` templates while enabling contextual roasts when media is playing. |
| **Activity timestamp (`last_active`)** | Recursive mtime scan (excluding `.git/`) | Captures non-committed file edits across all subdirectories accurately without git command execution overhead. |
