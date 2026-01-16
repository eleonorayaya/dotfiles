#!/usr/bin/env zsh

# Sketchybar theme generator
# Usage: ./sketchybar.sh <theme_name> <theme_json_path>

THEME_NAME=$1
THEME_JSON=$2

if [ -z "$THEME_NAME" ] || [ -z "$THEME_JSON" ]; then
  echo "Error: Theme name and JSON path required"
  echo "Usage: $0 <theme_name> <theme_json_path>"
  exit 1
fi

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

# Color conversion function: #RRGGBB -> 0xRRGGBBAA
# Args: color (hex), opacity (0.0-1.0, default 1.0)
convert_color() {
  local color=$1
  local opacity=${2:-1.0}

  # Remove # prefix
  color=${color#\#}

  # Convert opacity to hex (0.0-1.0 -> 00-FF)
  local alpha_decimal=$(printf "%.0f" $(echo "$opacity * 255" | bc))
  local alpha_hex=$(printf "%02X" $alpha_decimal)

  # Append alpha and convert to uppercase with 0x prefix
  echo "0x${alpha_hex}${(U)color}"
}

# Extract colors from JSON
SURFACE=$(jq -r '.colors.surface' "$THEME_JSON")
SURFACE_BORDER=$(jq -r '.colors.surfaceBorder' "$THEME_JSON")
TEXT_ON_SURFACE=$(jq -r '.colors.textOnSurface' "$THEME_JSON")
ACCENT=$(jq -r '.colors.accent' "$THEME_JSON")

# Convert colors to sketchybar format
# Bar gets 0.85 opacity, everything else is fully opaque
BAR_COLOR=$(convert_color "$SURFACE" 0.85)
BAR_BORDER_COLOR=$(convert_color "$SURFACE_BORDER")
POPUP_BACKGROUND_COLOR=$(convert_color "$SURFACE")
POPUP_BORDER_COLOR=$(convert_color "$SURFACE_BORDER")
SPACES_WRAPPER_BACKGROUND=$(convert_color "$SURFACE")

ICON_COLOR=$(convert_color "$TEXT_ON_SURFACE")
LABEL_COLOR=$(convert_color "$TEXT_ON_SURFACE")

ICON_HIGHLIGHT_COLOR=$(convert_color "$ACCENT")
LABEL_HIGHLIGHT_COLOR=$(convert_color "$ACCENT")
ACTIVE_WORKSPACE_COLOR=$(convert_color "$ACCENT")
SPACES_ITEM_BACKGROUND=$(convert_color "$ACCENT")

# Get script directory and dotfiles root
SCRIPT_DIR="${0:A:h}"
DOTFILES_ROOT="${SCRIPT_DIR:h:h}"

# Create output directory
OUTPUT_DIR="$DOTFILES_ROOT/sketchybar/themes"
mkdir -p "$OUTPUT_DIR"

# Generate theme file
OUTPUT_FILE="$OUTPUT_DIR/${THEME_NAME}.sh"

cat > "$OUTPUT_FILE" << EOF
#!/usr/bin/env zsh

export BAR_COLOR="$BAR_COLOR"
export BAR_BORDER_COLOR="$BAR_BORDER_COLOR"

export ICON_COLOR="$ICON_COLOR"
export ICON_HIGHLIGHT_COLOR="$ICON_HIGHLIGHT_COLOR"
export LABEL_COLOR="$LABEL_COLOR"
export LABEL_HIGHLIGHT_COLOR="$LABEL_HIGHLIGHT_COLOR"

export POPUP_BORDER_COLOR="$POPUP_BORDER_COLOR"
export POPUP_BACKGROUND_COLOR="$POPUP_BACKGROUND_COLOR"

export ACTIVE_WORKSPACE_COLOR="$ACTIVE_WORKSPACE_COLOR"

export SPACES_WRAPPER_BACKGROUND="$SPACES_WRAPPER_BACKGROUND"
export SPACES_ITEM_BACKGROUND="$SPACES_ITEM_BACKGROUND"
EOF

echo "[sketchybar] Generated theme: $OUTPUT_FILE"

