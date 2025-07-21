SKHD_CONFIG_DIR="$CONFIG_PATH/skhd"
SKHD_CONFIG=skhdrc
SKHD_CONFIG_OUT="$SKHD_CONFIG"

echo "[skhd] checking installation"
if ! command -v skhd>/dev/null 2>&1
then
  echo "[skhd] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[skhd] already installed"
fi

echo "[skhd] checking config"

mkdir -p $SKHD_CONFIG_DIR

if [ -f "$SKHD_CONFIG_DIR/$SKHD_CONFIG_OUT" ]; then
  echo "[skhd] config already linked"
else
  ln -s $DOTFILE_PATH/SKHD/$SKHD_CONFIG $SKHD_CONFIG_DIR/$SKHD_CONFIG_OUT
  echo "[skhd] linked config"
fi

# Start skhd at login
skhd --install-service>/dev/null

