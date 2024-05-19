-- ARGV[1]: 最高水位
-- ARGV[2]: 水流速度/秒
-- ARGV[3]: 当前时间（秒）

local peakLevel = tonumber(ARGV[1])
local currentVelocity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local lastTime = tonumber(redis.call("hget", KEYS[1], "lastTime"))
local currentLevel = tonumber(redis.call("hget", KEYS[1], "currentLevel"))
-- 初始化
if lastTime == nil then
	lastTime = now
	currentLevel = 0
	redis.call("hmset", KEYS[1], "currentLevel", currentLevel, "lastTime", lastTime)
end

-- 尝试放水
-- 距离上次放水的时间
local interval = now - lastTime
if interval > 0 then
	-- 当前水位-距离上次放水的时间(秒)*水流速度
	local newLevel = currentLevel - interval * currentVelocity
	if newLevel < 0 then
		newLevel = 0
	end
	currentLevel = newLevel
	redis.call("hmset", KEYS[1], "currentLevel", newLevel, "lastTime", now)
end

-- 若到达最高水位，请求失败
if currentLevel >= peakLevel then
	return 0
end
-- 若没有到达最高水位，当前水位+1，请求成功
redis.call("hincrby", KEYS[1], "currentLevel", 1)
redis.call("expire", KEYS[1], peakLevel / currentVelocity)
return 1