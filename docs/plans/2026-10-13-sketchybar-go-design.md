# Sketchybar in Go: design

Status: first cut implemented; styling and calendar deferred.

## Decisions

- **Language: pure Go daemon.** `shizuku bar run` is the sketchybar config. `sketchybarrc` sources `shizuku.zshenv` for `PATH` and `exec`s it. No cgo, so releases stay `CGO_ENABLED=0`.
- **Transport.**
  - Commands go to sketchybar through the `sketchybar` CLI, batched into one call per update.
  - Events come back through one hidden sink item (`shizuku.events`). Its script is `shizuku bar emit`, which forwards `NAME`/`SENDER`/`INFO` and the trigger vars to the daemon over a unix socket.
  - Bursts of queued events are coalesced, so each handler runs once per burst.
  - Routing is by item name. Sink events go to subscribed handlers. Any other name goes to the click handler that owns its prefix, and that handler runs off the event loop.
  - Only items with click handlers (Bluetooth) carry their own emit script, through `Batch.OnClick`.
  - Plain clicks use `click_script` directly.
  - Mach-port delivery (as SbarLua does it) could replace this later without touching items.
- **Config.**
  - `Generate()` writes `~/.config/sketchybar/bar.json`, holding the theme colors at `WindowOpacity` and the per-profile mail app. Layout values are literals in the daemon until styling is decided.
  - The daemon reads only that file and never loads the shizuku builder.
  - Per-profile behaviour comes from `sketchybar.New(Options{...})` overlays.
- **Lifecycle.**
  - The daemon writes a pidfile next to its socket. A new daemon that finds a live socket sends the old one SIGTERM (SIGKILL after 2s), so `sketchybar --reload` replaces the old daemon cleanly.
  - The daemon exits after three failed `sketchybar --query bar` health checks. The check is skipped when an update was applied recently.
  - Every command has a 10s timeout.

## Look: ribbon islands

Chosen from a moodboard of pastel and kawaii bars. Each group of items sits in its own floating rounded island: blush fill, rose outline, gaps between islands. The bar itself is transparent. Labels use M PLUS Rounded 1c ("Rounded Mplus 1c"). System icons are SF Symbols, and app icons come from sketchybar-app-font. Raspberry is reserved for the current workspace, the clock and badge bubbles.

| Island | Contents | Source / update |
|---|---|---|
| ♥ | heart (♡ when Bluetooth is off); click opens paired devices, click a device to connect/disconnect | `blueutil` |
| ✿ n | current AeroSpace workspace only, raspberry | aerospace triggers, `aerospace list-workspaces --focused` |
| front app | app-font icon + lowercase name | `front_app_switched` |
| kaomoji (center) | reacting face | 5s; priority sleepy (0–6h) > lost (offline) > flustered (CPU ≥ 80% twice) > sparkly (new unread, 5 min) > happy |
| Utena | `˚ ₊ session ₊ ˚` | `/sessions` + activation long-poll |
| badges | app-font icons with raspberry count bubbles; "•" renders as a dot; hidden at zero | `lsappinfo` Dock badge, 15s |
| network | Wi-Fi / globe / offline icon + ↓↑ rate | `route`, `ipconfig`, `netstat`, 3s |
| cpu · mem | SF Symbols chips, one bracketed island | `ps`, `vm_stat`, 5s |
| clock | raspberry: time + date | minute-aligned, `system_woke` |

Profiles set the badge apps: `work` uses Mail + Slack, and `personal` uses Proton Mail + Discord.

## Deferred

- **Visual style.** This cut is a port of the previous floating pill. The islands vs pill vs full-width decision is still open.
- **Meeting item.** It should appear 15 minutes before a meeting, stay until the meeting ends, and click to join. It needs an EventKit reader (a small Swift helper was proposed). The clock click opens Calendar for now.
- **Calendar popup** on the clock.
- **Wi-Fi SSID.** macOS redacts it without Location permission.
- **Bigger Utena integration.**
- **Per-workspace app icons** (sketchybar-app-font). Workspaces keep the fixed icon map for now.
