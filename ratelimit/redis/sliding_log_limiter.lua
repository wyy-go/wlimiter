-- ARGV[1]: 当前小窗口值
-- ARGV[2]: 第一个策略的窗口时间大小
-- ARGV[i * 2 + 1]: 每个策略的起始小窗口值
-- ARGV[i * 2 + 2]: 每个策略的窗口请求上限

local currentSmallWindow = tonumber(ARGV[1])
-- 第一个策略的窗口时间大小
local window = tonumber(ARGV[2])
-- 第一个策略的起始小窗口值
local startSmallWindow = tonumber(ARGV[3])
local strategiesLen = #(ARGV) / 2 - 1

-- 计算每个策略当前窗口的请求总数
local counters = redis.call("hgetall", KEYS[1])
local counts = {}
-- 初始化counts
for j = 1, strategiesLen do
	counts[j] = 0
end

for i = 1, #(counters) / 2 do
	local smallWindow = tonumber(counters[i * 2 - 1])
	local counter = tonumber(counters[i * 2])
	if smallWindow < startSmallWindow then
		redis.call("hdel", KEYS[1], smallWindow)
	else
		for j = 1, strategiesLen do
			if smallWindow >= tonumber(ARGV[j * 2 + 1]) then
				counts[j] = counts[j] + counter
			end
		end
	end
end

-- 若到达对应策略窗口请求上限，请求失败，返回违背的策略下标
for i = 1, strategiesLen do
	if counts[i] >= tonumber(ARGV[i * 2 + 2]) then
		return i - 1
	end
end

-- 若没到窗口请求上限，当前小窗口计数器+1，请求成功
redis.call("hincrby", KEYS[1], currentSmallWindow, 1)
redis.call("pexpire", KEYS[1], window)
return -1