echo "[desktoppr] checking installation"
if ! command -v desktoppr >/dev/null 2>&1
then
  echo "[desktoppr] was not properly installed via homebrew, exiting"
  exit 1
else
  echo "[desktoppr] already installed"
fi

desktoppr $DOTFILE_PATH/wallpaper/cozy-autumn-rain.png
