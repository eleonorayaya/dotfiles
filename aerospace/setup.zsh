AEROSPACE_CONFIG_DIR="$CONFIG_PATH/aerospace"
AEROSPACE_CONFIG=aerospace.toml

echo "[aerospace] checking installation"
if ! command -v aerospace >/dev/null 2>&1
then
  brew install --cask nikitabobko/tap/aerospace
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

