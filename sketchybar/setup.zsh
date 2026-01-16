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

# Create styles directory and symlink all style files
mkdir -p $SKETCHYBAR_CONFIG_DIR/styles

for style_file in $DOTFILE_PATH/sketchybar/styles/*; do
  if [ -f "$style_file" ]; then
    style_name=$(basename "$style_file")
    ln -sf $style_file $SKETCHYBAR_CONFIG_DIR/styles/$style_name
    echo "[sketchybar] linked style: $style_name"
  fi
done

# Create items directory and symlink all item files
mkdir -p $SKETCHYBAR_CONFIG_DIR/items

for item_file in $DOTFILE_PATH/sketchybar/items/*; do
  if [ -f "$item_file" ]; then
    item_name=$(basename "$item_file")
    chmod +x $item_file
    ln -sf $item_file $SKETCHYBAR_CONFIG_DIR/items/$item_name
    echo "[sketchybar] linked item: $item_name"
  fi
done

# Create plugins directory and symlink all plugin files
mkdir -p $SKETCHYBAR_CONFIG_DIR/plugins

for plugin_file in $DOTFILE_PATH/sketchybar/plugins/*; do
  if [ -f "$plugin_file" ]; then
    plugin_name=$(basename "$plugin_file")
    chmod +x $plugin_file
    ln -sf $plugin_file $SKETCHYBAR_CONFIG_DIR/plugins/$plugin_name
    echo "[sketchybar] linked plugin: $plugin_name"
  fi
done

# Start or restart sketchybar
if brew services list | grep sketchybar | grep -q started; then
  echo "[sketchybar] restarting service"
  brew services restart sketchybar
else
  echo "[sketchybar] starting service"
  brew services start sketchybar
fi

