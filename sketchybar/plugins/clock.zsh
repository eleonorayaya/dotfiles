#!/usr/bin/env zsh

ICON=$(date '+%a %B %d')
LABEL=$(date '+%I:%M %p')

clock=(
  label="${LABEL:-error}"
  icon="${ICON:-error}"
)

sketchybar --set "$NAME" "${clock[@]}"

