# Configuration & Data Files

`nudged` stores its configuration, registry, and state files under `~/.blackboxx/nudge/`.

---

## Configuration (`config.json`)

Located at `~/.blackboxx/nudge/config.json`. If missing at startup, `nudged` generates it with default values.

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

### Configuration Options

| Field | Default | Description |
|-------|---------|-------------|
| `scan_roots` | `[]` | List of root directories to scan for Git projects (e.g. `["/home/spectre/Projects"]`). |
| `ignored_paths` | `[]` | List of directories to skip during scanning (e.g. vendor, node_modules parent directories). |
| `idle_threshold_minutes` | `20` | Minutes of user inactivity before the initial idle nudge fires. |
| `idle_repeat_interval_minutes` | `20` | Interval between repeated idle nudges while the user remains inactive. |
| `idle_poll_interval_seconds` | `30` | Interval at which D-Bus is queried for user idle time. |
| `dormant_threshold_days` | `7` | Days of inactivity before a project is flagged as dormant. |
| `project_scan_interval_minutes` | `15` | Interval between project directory discovery scans. |
| `spectretts_socket_path` | `/tmp/spectretts.sock` | Path to the SpectreTTS Unix domain socket. |

> [!IMPORTANT]
> `nudged` does not automatically reload `config.json` while running. Restart the service after modifying this file:
> ```sh
> systemctl --user restart nudged
> ```

---

## Data Files

| File | Purpose | Mutated By |
|------|---------|------------|
| `config.json` | User settings, thresholds, and scan roots | User / `nudged -setup` |
| `projects.json` | Registry of discovered and manual projects | Daemon / User |
| `state.json` | Last-nudged timestamps and cooldown tracking | Daemon |

---

## Adding Projects Manually to `projects.json`

The daemon never deletes entries from `projects.json` if a directory is missing from a scan. You can manually register or annotate projects:

```json
{
  "my-custom-project": {
    "path": "/home/spectre/Projects/my-custom-project",
    "summary": "Core backend API service",
    "source": "manual",
    "last_active": "2026-09-01T12:00:00Z"
  }
}
```

- **`source`**: Set to `"manual"` for custom entries. `summary` will never be modified by the scanner.
- **`last_active`**: ISO 8601 timestamp. Updated automatically if the project path is within `scan_roots`.
- **Key Name**: The JSON object key (e.g. `"my-custom-project"`) serves as the display name spoken by the TTS engine.

> [!TIP]
> Stop `nudged` before manual edits to `projects.json` to prevent race conditions during write cycles:
> ```sh
> systemctl --user stop nudged
> $EDITOR ~/.blackboxx/nudge/projects.json
> systemctl --user start nudged
> ```
