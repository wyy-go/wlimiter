package redis

import _ "embed"

//go:embed fixed_window_limiter.lua
var FixedWindowLimitScript string

//go:embed leaky_bucket_limiter.lua
var LeakyBucketLimitScript string

//go:embed sliding_log_limiter.lua
var SlidingLogLimitScript string

//go:embed sliding_window_limiter_hash.lua
var SlidingWindowLimiterHashScript string

//go:embed sliding_window_limiter_list.lua
var SlidingWindowLimiterListScript string

//go:embed token_bucket_limiter.lua
var TokenBucketLimiterScript string
