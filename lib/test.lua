-- Read input from stdin line by line using io.lines()
local function trim(s)
    return s:match("^%s*(.-)%s*$")
end

o = io.read()
print("Input data is: ", o)
print("Hello, World!")