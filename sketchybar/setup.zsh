SKETCHYBAR_CONFIG_DIR="$CONFIG_PATH/sketchybar"
SKETCHYBAR_CONFIG=sketchybarrc.sh
SKETCHYBAR_CONFIG_OUT="sketchybarrc"

echo "[sketchybar] checking installation"
if ! command -v sketchybar >/dev/null 2>&1
then
  echo "[sketchybar] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[sketchybar] already installed"
fi

echo "[sketchybar] checking config"

mkdir -p $SKETCHYBAR_CONFIG_DIR

if [ -L "$SKETCHYBAR_CONFIG_DIR/$SKETCHYBAR_CONFIG_OUT" ]; then
  echo "[sketchybar] config already linked"
else
  ln -s $DOTFILE_PATH/sketchybar/$SKETCHYBAR_CONFIG $SKETCHYBAR_CONFIG_DIR/$SKETCHYBAR_CONFIG_OUT
  echo "[sketchybar] linked config"
fi

# Start sketchybar at login
brew services start sketchybar

