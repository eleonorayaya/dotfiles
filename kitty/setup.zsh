KITTY_CONFIG_DIR="$CONFIG_PATH/kitty"
KITTY_CONFIG=kitty.conf
THEME_CONFIG=monade.conf

# TODO: kitty isn't being added to the PATH, so this check passes even if kitty isn't installed
echo "[kitty] checking installation"
if [ -d "/Applications/kitty.app/" ];
then
  echo "[kitty] already installed"
else
  curl -L https://sw.kovidgoyal.net/kitty/installer.sh | sh /dev/stdin
fi

echo "[kitty] checking config"

mkdir -p $KITTY_CONFIG_DIR

if [ -L "$KITTY_CONFIG_DIR/$KITTY_CONFIG" ]; then
  echo "[kitty] config already linked"
else
  ln -s $DOTFILE_PATH/kitty/$KITTY_CONFIG $KITTY_CONFIG_DIR/$KITTY_CONFIG
  echo "[kitty] linked config"
fi

echo "[kitty] checking theme"
if [ -L "$KITTY_CONFIG_DIR/$THEME_CONFIG" ]; then
  echo "[kitty] theme already linked"
else
  ln -s $DOTFILE_PATH/kitty/$THEME_CONFIG $KITTY_CONFIG_DIR/$THEME_CONFIG
  echo "[kitty] linked theme"
fi

