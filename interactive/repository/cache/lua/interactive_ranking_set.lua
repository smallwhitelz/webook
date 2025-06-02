local zsetName = KEYS[1]                     -- 有序集合的名称
local memberToIncrement = ARGV[1]            -- 要增加分数的元素
local specifiedScore = tonumber(ARGV[2])     -- 指定的分数，如果未指定则默认为1

-- 检查有序集合是否存在，如果不存在则创建并设置指定元素的分数
local exists = redis.call("EXISTS", zsetName)
if exists == 0 then
    redis.call("ZADD", zsetName, specifiedScore, memberToIncrement)
    return specifiedScore
end

-- 获取指定元素的当前分数
local currentScore = redis.call("ZSCORE", zsetName, memberToIncrement)

if currentScore then
    -- 如果元素存在，将分数加1
    local newScore = currentScore + 1
    redis.call("ZADD", zsetName, newScore, memberToIncrement)
    return newScore
else
    -- 如果元素不存在，将分数设置为指定的分数
    redis.call("ZADD", zsetName, specifiedScore, memberToIncrement)
    return specifiedScore
end