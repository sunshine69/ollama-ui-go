local input = os.getenv("INPUT_DATA_FILE")

local file = io.open(input, "r")
if file then
    local content = file:read("*all")
    file:close()
    print(content)
else
    print("Could not open file: " .. input)
end