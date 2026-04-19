#!/bin/bash
# wrk Lua脚本压测 - 支持自定义请求模式
# wrk支持Lua脚本来自定义请求、响应处理等

set -e

# 配置
WRK_URL="${WRK_URL:-http://localhost:8000}"
DURATION="${DURATION:-30s"
THREADS="${THREADS:-12}"
CONNECTIONS="${CONNECTIONS:-400}"

#wrk配置
wrk -t$THREADS -c$CONNECTIONS -d$DURATION \
  -s /home/wood/Desktop/Distributed-Software-Course-Project/benchmark/post_test.lua \
  $WRK_URL/api/v1/users/login

-- post_test.lua
-- 自定义POST请求的Lua脚本

wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.body = '{"username":"testuser","password":"test123"}'

-- 请求计数器
local counter = 0

request = function()
    counter = counter + 1
    -- 可以动态修改body
    return wrk.format(nil, nil, nil, '{"username":"user' .. counter .. '","password":"pass' .. counter .. '"}')
end

response = function(status, headers, body)
    if status == 429 then
        io.write("R")  -- 标记限流响应
    elseif status >= 500 then
        io.write("E")  -- 标记错误响应
    end
end

done = function(summary)
    io.write("\n")
    io.write("Total requests: " .. summary.requests .. "\n")
    io.write("Errors (5xx): " .. summary.errors .. "\n")
end