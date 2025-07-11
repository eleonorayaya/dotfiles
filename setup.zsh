# # Fonts
# brew install --cask font-meslo-lg-nerd-font
#
# # Tools
# brew install fzf
# brew install lsd
# brew install bat
# npm install --global fkill-cli
#
# # Languages / Runtimes
# brew install rust
#
# # Neovim
# brew install neovim
# brew install ripgrep
#
# # Link Config Files
# mkdir -p ~/.config
#
# ln -s ~/.dotfiles/nvim ~/.config/nvim
# ln -s ~/.dotfiles/.tmux.conf ~/.tmux.conf
#
# # NVM setup
# mkdir -p ~/.nvm
# ln -s ~/.dotfiles/nvm-default-packages ~/.nvm/default-packages
#
# # Install tmux sessionizer
# # Ensure that ~/.cargo/bin is in your PATH (add to .env file)
# cargo install tmux-sessionizer
#
# git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm
# # Run C-b I to install plugins
#

DOTFILE_PATH=~/.dotfiles
CONFIG_PATH=~/.config

# TODO: kitty isn't being added to the PATH, so this check passes even if kitty isn't installed
echo "checking if kitty is installed"
if ! command -v kitty >/dev/null 2>&1
then
  curl -L https://sw.kovidgoyal.net/kitty/installer.sh | sh /dev/stdin
else
  echo "kitty is already installed"
fi

# TODO: ensure kitty config directory exists
echo "linking kitty config"
if [ -f "$CONFIG_PATH/kitty/kitty.conf" ]; then
  echo "Kitty config already linked"
else
  ln -s $DOTFILE_PATH/kitty/kitty.conf $CONFIG_PATH/kitty/kitty.conf
fi

if [ -f "$CONFIG_PATH/kitty/rose-pine-moon.conf" ]; then
  echo "Kitty theme already linked"
else
  ln -s $DOTFILE_PATH/kitty/rose-pine-moon.conf $CONFIG_PATH/kitty/rose-pine-moon.conf
fi

