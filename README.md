# Dotfiles

## Env Config

When first setting up dotfiles, create a new file `env.zsh` for machine-specific config. This file will not be tracked by git.

The nvim configs here reference some environment variables that can be set in this file (or anywhere). See `nvim/lua/core/env_config.lua` for supported variables.

## Generating fastfetch image
Convert the image with chafa and write it to a file
```
 chafa --size 30x20 --align top,left -f symbols --symbols block <image_name>.png > logo.data
```

## Helpful Links

### Nvim

- [Lua Cheatsheet](https://devhints.io/lua)
- [Nvim Lua Guide](https://github.com/nanotee/nvim-lua-guide)

