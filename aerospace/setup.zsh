Aerospace requires disabling native spaces - run this and log out and back in
# defaults write com.apple.spaces spans-displays -bool true && killall SystemUIServer

AEROSPACE_CONFIG_DIR="$CONFIG_PATH/aerospace"
AEROSPACE_CONFIG=aerospace.toml

echo "[aerospace] checking installation"
if ! command -v aerospace >/dev/null 2>&1
then
  echo "[aerospace] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[aerospace] already installed"
fi

echo "[aerospace] checking config"

mkdir -p $AEROSPACE_CONFIG_DIR

if [ -f "$AEROSPACE_CONFIG_DIR/$AEROSPACE_CONFIG" ]; then
  echo "[aerospace] config already linked"
else
  ln -s $DOTFILE_PATH/aerospace/$AEROSPACE_CONFIG $AEROSPACE_CONFIG_DIR/$AEROSPACE_CONFIG
  echo "[aerospace] linked config"
fi

