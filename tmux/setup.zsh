TMUX_CONFIG_DIR="$CONFIG_PATH/tmux"
TMUX_CONFIG=tmux.conf
TMUX_PLUGIN_DIR=~/.tmux/plugins

echo "[tmux] checking installation"
if ! command -v tmux >/dev/null 2>&1
then
  echo "[tmux] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[tmux] already installed"
fi

echo "[tmux] checking tms installation"
if ! command -v tms >/dev/null 2>&1
then
  cargo install tmux-sessionizer
  exit 1
else
  echo "[tmux] already installed"
fi

echo "[tmux] checking config"

mkdir -p $TMUX_CONFIG_DIR

if [ -f "$TMUX_CONFIG_DIR/$TMUX_CONFIG" ]; then
  echo "[tmux] config already linked"
else
  ln -s $DOTFILE_PATH/tmux/$TMUX_CONFIG $TMUX_CONFIG_DIR/$TMUX_CONFIG
  echo "[tmux] linked config"
fi

mkdir -p $TMUX_PLUGIN_DIR

echo "[tmux] checking tpm installation"

if [ -d "$TMUX_PLUGIN_DIR/tpm/" ]; then
  echo "[tmux] tpm already installed"
else
  git clone https://github.com/tmux-plugins/tpm $TMUX_PLUGIN_DIR/tpm
fi
