json = require("json")
http = require("http")
yaml = require("yaml")

local input_data_file = os.getenv("INPUT_DATA_FILE")
local ollama_host = os.getenv("OLLAMA_HOST")
local file = io.open(input_data_file, "r")
local input_data = file:read("*all")
file:close()
local success, decoded_args = pcall(json:decode, input_data)
if not success then
    print("Error json decoding: " .. tostring(input_data))
    decoded_args = input_data
-- else
--     print("Decoded: " .. decoded_args)
end
-- print(ollama_host)
print(decoded_args)

function get_current_weather(decoded_args)
    local url = "http://api.openweathermap.org/data/2.5/weather?q=" .. decoded_args.city .. "&appid=" .. decoded_args.api_key
    local response_body = {}
    local res, code, response_headers, status = http.request {
        url = url,
        method = "GET",
        sink = ltn12.sink.table(response_body)
    }
    if code == 200 then
        local response = table.concat(response_body)
        local decoded_response = json.decode(response)
        return decoded_response
    else
        return "Error: " .. code
    end
end