package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/wyy-go/wlimiter/ratelimit"
	redisRateLimit "github.com/wyy-go/wlimiter/ratelimit/redis"
	"time"
)

// TokenBucketLimiter 令牌桶限流器
type tokenBucketLimiter struct {
	capacity int           // 容量
	rate     int           // 发放令牌速率/秒
	client   *redis.Client // Redis客户端
	script   *redis.Script // TryAcquire脚本
}

func NewTokenBucketLimiter(client *redis.Client, capacity, rate int) ratelimit.RedisLimiter {
	return &tokenBucketLimiter{
		capacity: capacity,
		rate:     rate,
		client:   client,
		script:   redis.NewScript(redisRateLimit.TokenBucketLimiterScript),
	}
}

func (l *tokenBucketLimiter) TryAcquire(ctx context.Context, resource string) error {
	// 当前时间
	now := time.Now().Unix()
	success, err := l.script.Run(ctx, l.client, []string{resource}, l.capacity, l.rate, now).Bool()
	if err != nil {
		return err
	}
	// 若到达窗口请求上限，请求失败
	if !success {
		return redisRateLimit.ErrAcquireFailed
	}
	return nil
}
