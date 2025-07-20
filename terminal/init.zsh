source $(brew --prefix)/share/antigen/antigen.zsh

antigen bundle jeffreytse/zsh-vi-mode > /dev/null

antigen apply

. ~/.dotfiles/terminal/alias.zsh
. ~/.dotfiles/terminal/functions.zsh
. ~/.dotfiles/terminal/git-autocomplete.zsh

eval "$(oh-my-posh init zsh --config ~/.dotfiles/terminal/ohmyposh.json)"
