#! /usr/bin/env zsh
FASTFETCH_CONFIG_DIR="$CONFIG_PATH/fastfetch"
FASTFETCH_CONFIG="config.jsonc"

echo "[fastfetch] checking installation"
if ! command -v fastfetch >/dev/null 2>&1
then
  echo "[fastfetch] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[fastfetch] already installed"
fi

echo "[fastfetch] checking config"

mkdir -p $FASTFETCH_CONFIG_DIR

if [ -L "$FASTFETCH_CONFIG_DIR/$FASTFETCH_CONFIG" ]; then
  echo "[fastfetch] config already linked"
else
  ln -s $DOTFILE_PATH/fastfetch/$FASTFETCH_CONFIG $FASTFETCH_CONFIG_DIR/$FASTFETCH_CONFIG
  echo "[fastfetch] linked config"
fi
