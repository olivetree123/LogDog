
function myHandler(dataBytes)
    package.path = "/Users/gao/code/gowork/src/logDog/*.so"
    local json = require("cjson")
    local data = json.decode(dataBytes)
    --local add_data = data["add_data"]
    --local message = data["message"]
    --for key, val in pairs(message) do
    --    print(key, val)
    --end
    return json.encode(data)
end