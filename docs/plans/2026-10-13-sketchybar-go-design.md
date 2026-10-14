# Sketchybar in Go: design

Status: implemented; calendar/meetings deferred.

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
  - `Generate()` writes `~/.config/sketchybar/bar.json`, holding the theme colors at `WindowOpacity`, the per-profile badge apps, and the window-gap margins.
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
| ♥ (left) | decorative heart | — |
| ✿ n | current AeroSpace workspace only, raspberry | aerospace triggers, `aerospace list-workspaces --focused` |
| front app | app-font icon + lowercase name | `front_app_switched` |
| Utena (left) | `˚ ₊ session ₊ ˚` | `/sessions` + activation long-poll |
| kaomoji (center) | reacting face | 5s; priority sleepy (0–6h) > lost (offline) > flustered (CPU ≥ 80% twice) > sparkly (new unread, 5 min) > happy |
| badges | app-font icons with raspberry count bubbles; "•" renders as a dot; hidden at zero | `lsappinfo` Dock badge, 15s |
| connect | Wi-Fi / wired / offline icon + Bluetooth icon (faded when off); Bluetooth click opens paired devices | `route`, `ipconfig`, `blueutil`, 5s |
| cpu · mem | `cpu` label + pink sparkline with a twinkling ✦; `mem` label + meter with a heart knob | `ps`, `vm_stat`, 3s |
| clock | raspberry: time + date | minute-aligned, `system_woke` |

The bar's side margin follows `styles.Gaps.OuterLeft` (and the `GapOverride` when every monitor matches its pattern), so island edges line up with window edges. Islands are 12pt apart.

Profiles set the badge apps: `work` uses Mail + Slack, and `personal` uses Proton Mail + Discord.

## Deferred

- **Per-monitor margins.** Sketchybar has one margin for all displays, so mixed built-in + external setups use the default gap everywhere.
- **Meeting item.** It should appear 15 minutes before a meeting, stay until the meeting ends, and click to join. It needs an EventKit reader (a small Swift helper was proposed). The clock click opens Calendar for now.
- **Calendar popup** on the clock.
- **Wi-Fi SSID.** macOS redacts it without Location permission.
- **Bigger Utena integration.**
- **Per-workspace app icons** (sketchybar-app-font). Workspaces keep the fixed icon map for now.
