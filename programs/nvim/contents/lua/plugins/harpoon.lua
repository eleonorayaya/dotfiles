return {
  {
    "ThePrimeagen/harpoon",
    branch = "harpoon2",
    dependencies = { "nvim-lua/plenary.nvim" },
    config = function()
      local harpoon = require("harpoon")
      harpoon:setup()

      vim.keymap.set("n", "<leader>m", function() harpoon:list():add() end)
      vim.keymap.set("n", "<C-o>", function()
        local list = harpoon:list()

        Snacks.picker({
          title = "Harpoon",
          format = "file",
          finder = function()
            local items = {}
            for i = 1, list:length() do
              local item = list:get(i)
              if item then
                table.insert(items, {
                  text = item.value,
                  file = item.value,
                  idx = i,
                })
              end
            end
            return items
          end,
          confirm = function(picker, selected)
            picker:close()
            if selected then
              list:select(selected.idx)
            end
          end,
          win = {
            input = {
              keys = {
                ["<C-d>"] = { "harpoon_remove", mode = { "n", "i" } },
              },
            },
          },
          actions = {
            harpoon_remove = function(picker, selected)
              if not selected then return end
              list:remove_at(selected.idx)
              picker:refresh()
            end,
          },
        })
      end)
    end,
  },
}
