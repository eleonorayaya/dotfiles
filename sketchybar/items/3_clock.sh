#!/usr/bin/env bash

set -x

CLOCK_FONT="$FONT:Thin:16.0"

clock=(
  label.font="$FONT:Regular:16.0"
  label.padding_left=9
  icon.font="$CLOCK_FONT"
  update_freq=15 
  script="plugin_clock"
)

sketchybar --add item clock right \
  --set clock "${clock[@]}"

FONT="$FONT" NAME="clock" plugin_clock &
