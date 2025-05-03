return {
	"cordx56/rustowl",
	version = "*", -- Latest stable version
	build = "cargo install --path . --locked",
	lazy = false, -- This plugin is already lazy
	opts = {
		client = {
			on_attach = function(_, buffer)
				vim.keymap.set("n", "<leader>o", function()
					require("rustowl").toggle(buffer)
				end, { buffer = buffer, desc = "Toggle RustOwl" })
			end,
		},
	},
}
