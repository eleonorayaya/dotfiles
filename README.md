# Shizuku

A Go-based configuration management tool for dotfiles.

## Overview

Shizuku manages application configurations by generating files from templates, downloading remote resources, and syncing everything to the appropriate destinations.

## Servers

The base profile is a minimal, server-safe set: git, zsh (vi-mode + oh-my-posh), helix, bat, lsd, tuios. On Debian/Ubuntu it installs via apt, falling back to GitHub release binaries in `~/.local/bin`.

```sh
mkdir -p ~/.local/bin
curl -fsSL https://github.com/eleonorayaya/dotfiles/releases/latest/download/shizuku_linux_$(dpkg --print-architecture).tar.gz \
  | tar -xz -C ~/.local/bin shizuku
~/.local/bin/shizuku install && ~/.local/bin/shizuku sync
```

Update later with `shizuku upgrade` (downloads the latest release when there is no source checkout).

On a desktop, set a profile: `shizuku config set profile personal` (or `work`). Both extend the `desktop` profile.
