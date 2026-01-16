#!/usr/bin/env zsh

# Set sketchybar theme
# Usage: ./set_theme.sh <theme_name>

THEME_NAME=$1

if [ -z "$THEME_NAME" ]; then
  echo "Usage: $0 <theme_name>"
  echo ""
  echo "Example:"
  echo "  $0 monade"
  exit 1
fi

# Get script directory and dotfiles root
SCRIPT_DIR="${0:A:h}"
DOTFILES_ROOT="${SCRIPT_DIR:h}"

# Set paths
THEME_FILE="$DOTFILES_ROOT/sketchybar/themes/${THEME_NAME}.sh"
CONFIG_DIR="${CONFIG_PATH:-$HOME/.config}/sketchybar"
THEME_LINK="$CONFIG_DIR/styles/theme.sh"

# Validate theme exists
if [ ! -f "$THEME_FILE" ]; then
  echo "[sketchybar] Error: Theme not found: $THEME_FILE"
  echo "[sketchybar] Generate it first with: ./themes/generate themes/${THEME_NAME}.json"
  exit 1
fi

# Create styles directory if it doesn't exist
mkdir -p "$CONFIG_DIR/styles"

# Create symlink
ln -sf "$THEME_FILE" "$THEME_LINK"
echo "[sketchybar] Theme set to: $THEME_NAME"

# Restart sketchybar if it's running
if brew services list | grep sketchybar | grep -q started; then
  echo "[sketchybar] Restarting to apply theme..."
  brew services restart sketchybar
fi
