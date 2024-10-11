# Dotfiles

## Env Config

When first setting up dotfiles, create a new file `env.zsh` for machine-specific config. This file will not be tracked by git.

The nvim configs here reference some environment variables that can be set in this file (or anywhere). See `nvim/lua/core/env_config.lua` for supported variables.
