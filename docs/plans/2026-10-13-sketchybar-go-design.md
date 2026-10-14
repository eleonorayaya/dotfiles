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

## Bottom shelf

The Dock is hidden for good (a separate `dock` app sets autohide and a 1000s reveal delay). A second sketchybar instance replaces it: `~/.local/bin/bottombar` is a symlink to sketchybar, because sketchybar names its instance, lock file and IPC service after the binary. A LaunchAgent (`sh.shizuku.bottombar`, logging to `~/Library/Logs/shizuku-bottombar.log`) keeps it running. Its rc execs `shizuku bar run --bar bottom`, a second daemon with its own socket. Sync reloads it in place and only re-bootstraps the agent when the plist changes.

| Island | Contents | Source / update |
|---|---|---|
| sessions (left) | live Utena sessions (active/creating/pending/completed); one pip per Claude agent: ● working, ○ idle/done, ✦ ready for review or needs attention; names truncated, overflow shown as `+N` | `GET /sessions`, 3s; click → `PUT /sessions/{id}/activate`; hover → outline + agent popup; bounce when an agent finishes |
| pet's room (center) | pixel Sylveon rendered to PNG by the daemon: blinks, strolls with a walk cycle, naps (idle > 5 min or 0–6h), hops when an agent finishes, waves when one needs attention | 1s tick; `ioreg` idle check every 10s |
| workspaces (right) | tiles for the workspaces on that monitor with app icons; each monitor's visible space highlighted | aerospace triggers; click → `aerospace workspace`; hover tint |

The shelf is centered in the bottom window gap (`styles.Gaps.OuterBottom`, now 64pt). Sketchybar applies a bottom bar's `y_offset` twice, so the offset is `(gap − height) / 4`. It hides on monitors whose gap override is too small to fit it (the built-in display).

## Deferred

- **Per-monitor margins.** Sketchybar has one margin for all displays, so mixed built-in + external setups use the default gap everywhere.
- **Meeting item.** It should appear 15 minutes before a meeting, stay until the meeting ends, and click to join. It needs an EventKit reader (a small Swift helper was proposed). The clock click opens Calendar for now.
- **Calendar popup** on the clock.
- **Wi-Fi SSID.** macOS redacts it without Location permission.
- **Bigger Utena integration.**
- **Per-workspace app icons** (sketchybar-app-font). Workspaces keep the fixed icon map for now.
