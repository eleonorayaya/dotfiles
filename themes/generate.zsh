#!/usr/bin/env zsh

# Theme generator main script
# Usage: ./generate <theme.json>
# Generates theme files for all configured tools

THEME_NAME=$1

if [ -z "$THEME_NAME" ]; then
  echo "Usage: $0 <theme.json>"
  echo ""
  echo "Example:"
  echo "  $0 themes/monade.json"
  exit 1
fi

THEME_JSON="themes/$THEME_NAME.json"
if [ ! -f "$THEME_JSON" ]; then
  echo "Error: Theme JSON not found: $THEME_JSON"
  exit 1
fi

# Check if jq is installed
if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required but not installed"
  echo "Install with: brew install jq"
  exit 1
fi

# Extract theme name from JSON
THEME_NAME=$(jq -r '.name' "$THEME_JSON")

if [ -z "$THEME_NAME" ] || [ "$THEME_NAME" = "null" ]; then
  echo "Error: Theme JSON must have a 'name' field"
  exit 1
fi

echo "Generating theme: $THEME_NAME"
echo ""

# Get absolute path to theme JSON
THEME_JSON_ABS=$(realpath "$THEME_JSON")

# Get script directory
SCRIPT_DIR="${0:A:h}"

# List of available generators
GENERATORS=(sketchybar)

# Run all generators
for generator in "${GENERATORS[@]}"; do
  GENERATOR_SCRIPT="$SCRIPT_DIR/generators/${generator}.sh"

  if [ ! -f "$GENERATOR_SCRIPT" ]; then
    echo "Warning: Generator not found: $GENERATOR_SCRIPT"
    continue
  fi

  "$GENERATOR_SCRIPT" "$THEME_NAME" "$THEME_JSON_ABS"
done

echo ""
echo "Theme generation complete!"
echo ""
echo "To activate the theme for a tool, run:"
for generator in "${GENERATORS[@]}"; do
  echo "  ./${generator}/set_theme.sh $THEME_NAME"
done

