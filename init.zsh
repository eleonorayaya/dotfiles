# omzsh config
export ZSH="$HOME/.oh-my-zsh"

#ZSH_THEME="robbyrussell"
plugins=(git nvm yarn vscode zsh-autosuggestions)

# my dotfiles
. ~/.dotfiles/starship.zsh
. ~/.dotfiles/alias.zsh
. ~/.dotfiles/env.zsh
. ~/.dotfiles/functions.zsh
. ~/.dotfiles/git-autocomplete.zsh
. ~/.dotfiles/shim.zsh

neofetch
