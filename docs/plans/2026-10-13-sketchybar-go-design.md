# Sketchybar in Go: design

Status: first cut implemented; styling and calendar deferred.

## Decisions

- **Language: pure Go daemon.** `shizuku bar run` is the sketchybar config. `sketchybarrc` sets `PATH` and `exec`s it. No cgo, so releases stay `CGO_ENABLED=0`.
- **Transport.**
  - Commands go to sketchybar through the `sketchybar` CLI, batched into one call per update.
  - Events come back through one hidden sink item (`shizuku.events`). Its script is `shizuku bar emit`, which forwards `NAME`/`SENDER`/`INFO` and the trigger vars to the daemon over a unix socket.
  - Only items with click handlers (Bluetooth) carry their own emit script.
  - Plain clicks use `click_script` directly.
  - Mach-port delivery (as SbarLua does it) could replace this later without touching items.
- **Config.**
  - `Generate()` writes `~/.config/sketchybar/bar.json`, holding theme colors at `WindowOpacity`, layout, and per-profile options.
  - The daemon reads only that file and never loads the shizuku builder.
  - Per-profile behaviour comes from `sketchybar.New(Options{...})` overlays.
- **Lifecycle.**
  - A new daemon sends `shizuku.quit` to an existing socket and waits for it to go away before listening, so `sketchybar --reload` replaces the old daemon cleanly.
  - The daemon exits after three failed `sketchybar --query bar` health checks.

## Items

| Item | Source | Update |
|---|---|---|
| Workspaces | `aerospace list-workspaces --all --json`, `list-windows --all --json` | aerospace triggers, `front_app_switched`, `space_windows_change`, `display_change`, `system_woke` |
| Front app | `$INFO` from `front_app_switched` | event |
| Clock | Go `time` | 15s |
| Mail unread | `lsappinfo info -only StatusLabel <app>` (Dock badge) | 30s, hidden at 0 |
| Network | `route get default`, `ipconfig getsummary`, `netstat -ibn` | 3s, connection type + ↓/↑ rate |
| CPU / memory | `ps -A -o %cpu=`, `vm_stat`, `hw.memsize` | 5s |
| Utena session | `GET /sessions` (`is_attached`), long-poll `/tmux/activations/next` | push + 50s fallback |
| Bluetooth | `blueutil --power`, `--paired --format json`, `--connect/--disconnect` | 30s icon; click opens a device popup |

Profiles: `work` uses `MailApp: "Mail"` and `personal` uses `MailApp: "Proton Mail"`. Calendar for both will come through Calendar.app, with Proton subscribed there via its ICS link.

## Deferred

- **Visual style.** This cut is a port of the previous floating pill. The islands vs pill vs full-width decision is still open.
- **Meeting item.** It should appear 15 minutes before a meeting, stay until the meeting ends, and click to join. It needs an EventKit reader (a small Swift helper was proposed). The clock click opens Calendar for now.
- **Calendar popup** on the clock.
- **Wi-Fi SSID.** macOS redacts it without Location permission.
- **Bigger Utena integration.**
- **Per-workspace app icons** (sketchybar-app-font). Workspaces keep the fixed icon map for now.
