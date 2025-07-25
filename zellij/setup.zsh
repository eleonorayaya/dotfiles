ZELLIJ_CONFIG_DIR="$CONFIG_PATH/zellij"
ZELLIJ_CONFIG="zellij.kdl"
ZELLIJ_CONFIG_OUT="config.kdl"
THEME_CONFIG="rose-pine-moon.kdl"
LAYOUT_CONFIG="layout.kdl"

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

echo "[zelli] checking layouts"

mkdir -p $ZELLIJ_CONFIG_DIR/layouts

if [ -L "$ZELLIJ_CONFIG_DIR/layouts/$LAYOUT_CONFIG" ]; then
  echo "[zellij] layout already linked"
else
  ln -s $DOTFILE_PATH/zellij/$LAYOUT_CONFIG $ZELLIJ_CONFIG_DIR/layouts/default.kdl
  echo "[zellij] linked layout"
fi

