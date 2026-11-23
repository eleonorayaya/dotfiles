(require (only-in "helix/keymaps.scm" keymap))
(require "forest/forest.scm")

(keymap (global)
        (normal (space (e ":forest-open"))))
