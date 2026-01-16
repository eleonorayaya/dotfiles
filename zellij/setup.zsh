#! /usr/bin/env zsh
ZELLIJ_CONFIG_DIR="$CONFIG_PATH/zellij"
ZELLIJ_CONFIG="zellij.kdl"
ZELLIJ_CONFIG_OUT="config.kdl"
THEME_CONFIG="monade.kdl"
LAYOUT_CONFIG="layout.kdl"
PLUGIN_CONFIG_DIR="$ZELLIJ_CONFIG_DIR/plugins"

declare -A plugins=(
  [vim-zellij-navigator]="https://github.com/hiasr/vim-zellij-navigator/releases/latest/download/vim-zellij-navigator.wasm"
  [zjstatus]="https://github.com/dj95/zjstatus/releases/latest/download/zjstatus.wasm"
)

echo "[zellij] checking installation"
if ! command -v zellij >/dev/null 2>&1
then
  echo "[zellij] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[zellij] already installed"
fi

echo "[zellij] checking config"

mkdir -p $ZELLIJ_CONFIG_DIR

if [ -L "$ZELLIJ_CONFIG_DIR/$ZELLIJ_CONFIG_OUT" ]; then
  echo "[zellij] config already linked"
else
  ln -s $DOTFILE_PATH/zellij/$ZELLIJ_CONFIG $ZELLIJ_CONFIG_DIR/$ZELLIJ_CONFIG_OUT
  echo "[zellij] linked config"
fi


echo "[zellij] checking theme"

mkdir -p $ZELLIJ_CONFIG_DIR/themes

if [ -f "$ZELLIJ_CONFIG_DIR/themes/$THEME_CONFIG" ]; then
  echo "[zellij] theme already linked"
else
  cp $DOTFILE_PATH/zellij/$THEME_CONFIG $ZELLIJ_CONFIG_DIR/themes/
  echo "[zellij] linked theme"
fi

echo "[zellij] checking layouts"

mkdir -p $ZELLIJ_CONFIG_DIR/layouts

if [ -L "$ZELLIJ_CONFIG_DIR/layouts/default.kdl" ]; then
  echo "[zellij] layout already linked"
else
  ln -s $DOTFILE_PATH/zellij/$LAYOUT_CONFIG $ZELLIJ_CONFIG_DIR/layouts/default.kdl
  echo "[zellij] linked layout"
fi

echo "[zellij] checking plugins"

mkdir -p $PLUGIN_CONFIG_DIR

for plugin_name plugin_path in ${(kv)plugins}
do
  echo "[zellij] checking plugin: $plugin_name"

  if [ -f "$PLUGIN_CONFIG_DIR/${plugin_name}.wasm" ]; then
    echo "[zellij] $plugin_name already installed"
  else
    curl -L $plugin_path > "$PLUGIN_CONFIG_DIR/${plugin_name}.wasm"
    echo "[zellij] $plugin_name installed"
  fi
done

