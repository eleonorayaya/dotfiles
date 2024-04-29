# Create symlinks so vscode uses our shared settings

SETTINGS_FILE=~/Library/Application\ Support/Code/User/settings.json
KEYBINDINGS_FILE=~/Library/Application\ Support/Code/User/keybindings.json

if [ -f $SETTINGS_FILE ]; then
  mv $SETTINGS_FILE $SETTINGS_FILE.bak
fi

if [ -f $KEYBINDINGS_FILE ]; then
  mv $KEYBINDINGS_FILE $KEYBINDINGS_FILE.bak
fi

ln -s ~/.dotfiles/vscode/settings.json $SETTINGS_FILE
ln -s ~/.dotfiles/vscode/keybindings.json $KEYBINDINGS_FILE
