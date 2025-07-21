NVIM_CONFIG_DIR="$CONFIG_PATH/nvim"

echo "[nvim] checking installation"
if ! command -v nvim >/dev/null 2>&1
then
  echo "[nvim] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[nvim] already installed"
fi

echo "[nvim] checking config"

if [ -L "$NVIM_CONFIG_DIR" ]; then
  echo "[nvim] config already linked"
else
  ln -s $DOTFILE_PATH/nvim $NVIM_CONFIG_DIR
  echo "[nvim] linked config"
fi

