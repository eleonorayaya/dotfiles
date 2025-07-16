# . ~/.dotfiles/omzsh.zsh

. ~/.dotfiles/alias.zsh
. ~/.dotfiles/env.zsh
. ~/.dotfiles/shared_env.zsh
. ~/.dotfiles/functions.zsh
. ~/.dotfiles/git-autocomplete.zsh

source $(brew --prefix)/opt/zsh-vi-mode/share/zsh-vi-mode/zsh-vi-mode.plugin.zsh
eval "$(oh-my-posh init zsh --config ~/.dotfiles/terminal/ohmyposh.toml)"

fastfetch
