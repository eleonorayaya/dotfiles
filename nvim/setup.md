## Link config

```
ln -s <absolute_path_to_dotfiles>/nvim ~/.config/nvim
```

## Packer

```
git clone --depth 1 https://github.com/wbthomason/packer.nvim\
 ~/.local/share/nvim/site/pack/packer/start/packer.nvim
```

## Keybindings

To add a Cmd+<key> command, the keymapping must first be done in iTerm settings

- Open iterm settings
- Select profiles > Hotkey window > Keys > Key Mappings
- Click the '+' button
- Enter shortcut
- For action, select 'Send text with vim special chars'
- For value to send, enter '\<M-key>' e.g. for Cmd+P enter '\<M-p>'
- In your vim keymap, map the input as '<M-key>'
