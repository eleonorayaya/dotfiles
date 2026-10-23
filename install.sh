#!/usr/bin/env bash
# Bootstraps a remote Linux workspace: fetches the shizuku release binary,
# then installs and syncs the coder profile.
set -euo pipefail

if [ "$(uname -s)" != "Linux" ]; then
  echo "install.sh only bootstraps Linux workspaces; skipping on $(uname -s)"
  exit 0
fi

bin_dir="$HOME/.local/bin"
shizuku="$bin_dir/shizuku"
mkdir -p "$bin_dir"

# SHIZUKU_BIN_URL points at a raw binary (e.g. a branch build); otherwise use the latest release.
if [ -n "${SHIZUKU_BIN_URL:-}" ]; then
  curl -fsSL "$SHIZUKU_BIN_URL" -o "$shizuku.tmp"
  chmod 0755 "$shizuku.tmp"
  mv "$shizuku.tmp" "$shizuku"
else
  arch="$(dpkg --print-architecture)"
  curl -fsSL "https://github.com/eleonorayaya/dotfiles/releases/latest/download/shizuku_linux_${arch}.tar.gz" \
    | tar -xz -C "$bin_dir" shizuku
fi

export PATH="$bin_dir:$PATH"
export SHIZUKU_PROFILE=coder
"$shizuku" config set profile coder

# A failed app install must not skip sync or the login shell, but the script
# still exits non-zero so the workspace shows the failure.
status=0
"$shizuku" install || status=$?
"$shizuku" sync

user="${USER:-$(id -un)}"
if zsh_path="$(command -v zsh)"; then
  current_shell="$(getent passwd "$user" | cut -d: -f7)"
  if [ "$current_shell" != "$zsh_path" ]; then
    if ! sudo -n chsh -s "$zsh_path" "$user"; then
      echo "could not change login shell to $zsh_path" >&2
      status=1
    fi
  fi
else
  echo "zsh is not installed; the login shell stays $(getent passwd "$user" | cut -d: -f7)" >&2
  status=1
fi

exit "$status"
