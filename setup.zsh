# Fonts
brew install --cask font-meslo-lg-nerd-font

# Neovim
brew install neovim
brew install ripgrep

mkdir -p ~/.config
ln -s ~/.dotfiles/nvim ~/.config/nvim
ln -s ~/.dotfiles/.tmux.conf ~/.tmux.conf

git clone https://github.com/tmux-plugins/tpm ~/.tmux/plugins/tpm
# Run C-b I to install plugins
