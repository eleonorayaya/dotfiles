JANKYBORDERS_CONFIG_DIR="$CONFIG_PATH/borders"
JANKYBORDERS_CONFIG=jankybordersrc.sh
JANKYBORDERS_CONFIG_OUT="bordersrc"

echo "[jankyborders] checking installation"
if ! command -v borders >/dev/null 2>&1
then
  echo "[jankyborders] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[jankyborders] already installed"
fi

echo "[jankyborders] checking config"

mkdir -p $JANKYBORDERS_CONFIG_DIR

if [ -L "$JANKYBORDERS_CONFIG_DIR/$JANKYBORDERS_CONFIG_OUT" ]; then
  echo "[jankyborders] config already linked"
else
  ln -s $DOTFILE_PATH/jankyborders/$JANKYBORDERS_CONFIG $JANKYBORDERS_CONFIG_DIR/$JANKYBORDERS_CONFIG_OUT
  echo "[jankyborders] linked config"
fi

# Start jankyborders at login
brew services start borders

