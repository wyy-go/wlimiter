package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/wyy-go/wlimiter/ratelimit"
	redisRateLimit "github.com/wyy-go/wlimiter/ratelimit/redis"
	"time"
)

// FixedWindowLimiter 固定窗口限流器
type fixedWindowLimiter struct {
	limit  int           // 窗口请求上限
	window int           // 窗口时间大小
	client *redis.Client // Redis客户端
	script *redis.Script // TryAcquire脚本
}

func NewFixedWindowLimiter(client *redis.Client, limit int, window time.Duration) (ratelimit.RedisLimiter, error) {
	// redis过期时间精度最大到毫秒，因此窗口必须能被毫秒整除
	if window%time.Millisecond != 0 {
		return nil, errors.New("the window uint must not be less than millisecond")
	}

	return &fixedWindowLimiter{
		limit:  limit,
		window: int(window / time.Millisecond),
		client: client,
		script: redis.NewScript(redisRateLimit.FixedWindowLimitScript),
	}, nil
}

func (l *fixedWindowLimiter) TryAcquire(ctx context.Context, resource string) error {
	success, err := l.script.Run(ctx, l.client, []string{resource}, l.window, l.limit).Bool()
	if err != nil {
		return err
	}
	// 若到达窗口请求上限，请求失败
	if !success {
		return redisRateLimit.ErrAcquireFailed
	}
	return nil
}
