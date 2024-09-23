vim.keymap.set("n", "<leader>ww", "<cmd>only<cr>", { desc = "Window: only" })
vim.keymap.set("n", "<leader>wW", "<cmd>enew<cr>", { desc = "Window: new" })
vim.keymap.set("n", "<leader>wd", "<cmd>close<cr>", { desc = "Window: close" })
vim.keymap.set("n", "<leader>wc", "<cmd>close<cr>", { desc = "Window: close" })

vim.keymap.set("n", "<leader>ws", "<cmd>split<cr>", { desc = "Window: split" })
vim.keymap.set("n", "<leader>wv", "<cmd>vsplit<cr>", { desc = "Window: vsplit" })
vim.keymap.set("n", "<M-k>", "<cmd>vsplit<cr>", { desc = "Window: vsplit" })

vim.keymap.set("n", "<leader>wh", "<cmd>wincmd h<cr>", { desc = "Window: move left" })
vim.keymap.set("n", "<leader>wl", "<cmd>wincmd l<cr>", { desc = "Window: move right" })
vim.keymap.set("n", "<leader>wj", "<cmd>wincmd j<cr>", { desc = "Window: move down" })
vim.keymap.set("n", "<leader>wk", "<cmd>wincmd k<cr>", { desc = "Window: move up" })
