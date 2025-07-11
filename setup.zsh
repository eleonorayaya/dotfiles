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

export DOTFILE_PATH=~/.dotfiles
export CONFIG_PATH=~/.config

./homebrew/setup.zsh

./aerospace/setup.zsh
./kitty/setup.zsh
