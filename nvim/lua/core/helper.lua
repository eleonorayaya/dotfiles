function _G.put(...)
	local objects = {}
	for i = 1, select("#", ...) do
		local v = select(i, ...)
		table.insert(objects, vim.inspect(v))
	end

	print(table.concat(objects, "\n"))
	return ...
end

function table.merge(t1, t2)
	for k, v in ipairs(t2) do
		table.insert(t1, v)
	end

	return t1
end
