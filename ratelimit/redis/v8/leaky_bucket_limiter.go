package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/wyy-go/wlimiter/ratelimit"
	redisRateLimit "github.com/wyy-go/wlimiter/ratelimit/redis"
	"time"
)

// LeakyBucketLimiter 漏桶限流器
type leakyBucketLimiter struct {
	peakLevel       int           // 最高水位
	currentVelocity int           // 水流速度/秒
	client          *redis.Client // Redis客户端
	script          *redis.Script // TryAcquire脚本
}

func NewLeakyBucketLimiter(client *redis.Client, peakLevel, currentVelocity int) ratelimit.RedisLimiter {
	return &leakyBucketLimiter{
		peakLevel:       peakLevel,
		currentVelocity: currentVelocity,
		client:          client,
		script:          redis.NewScript(redisRateLimit.LeakyBucketLimitScript),
	}
}

func (l *leakyBucketLimiter) TryAcquire(ctx context.Context, resource string) error {
	// 当前时间
	now := time.Now().Unix()
	success, err := l.script.Run(ctx, l.client, []string{resource}, l.peakLevel, l.currentVelocity, now).Bool()
	if err != nil {
		return err
	}
	// 若到达窗口请求上限，请求失败
	if !success {
		return redisRateLimit.ErrAcquireFailed
	}
	return nil
}
