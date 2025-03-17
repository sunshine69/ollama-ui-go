local input_data_file = os.getenv("INPUT_DATA_FILE")
local ollama_host = os.getenv("OLLAMA_HOST")
local file = io.open(input_data_file, "r")
local input_data = file:read("*all")
file:close()
local success, decoded_args = pcall(decode, input_data)
if not success then
    print("Error json decoding: " .. tostring(input_data))
else
    print("Decoded: " .. decoded_args)
end
print(ollama_host)
print(decoded_args)