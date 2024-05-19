package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/wyy-go/wlimiter/ratelimit"
	redisRateLimit "github.com/wyy-go/wlimiter/ratelimit/redis"
	"time"
)

// SlidingWindowLimiter 滑动窗口限流器
type slidingWindowLimiter struct {
	limit        int           // 窗口请求上限
	window       int64         // 窗口时间大小
	smallWindow  int64         // 小窗口时间大小
	smallWindows int64         // 小窗口数量
	client       *redis.Client // Redis客户端
	script       *redis.Script // TryAcquire脚本
}

func NewSlidingWindowLimiter(client *redis.Client, limit int, window, smallWindow time.Duration) (
	ratelimit.RedisLimiter, error) {
	// redis过期时间精度最大到毫秒，因此窗口必须能被毫秒整除
	if window%time.Millisecond != 0 || smallWindow%time.Millisecond != 0 {
		return nil, errors.New("the window uint must not be less than millisecond")
	}

	// 窗口时间必须能够被小窗口时间整除
	if window%smallWindow != 0 {
		return nil, errors.New("window cannot be split by integers")
	}

	return &slidingWindowLimiter{
		limit:        limit,
		window:       int64(window / time.Millisecond),
		smallWindow:  int64(smallWindow / time.Millisecond),
		smallWindows: int64(window / smallWindow),
		client:       client,
		script:       redis.NewScript(redisRateLimit.SlidingWindowLimiterListScript), // rateLimit.SlidingWindowLimiterHashScript
	}, nil
}

func (l *slidingWindowLimiter) TryAcquire(ctx context.Context, resource string) error {
	// 获取当前小窗口值
	currentSmallWindow := time.Now().UnixMilli() / l.smallWindow * l.smallWindow
	// 获取起始小窗口值
	startSmallWindow := currentSmallWindow - l.smallWindow*(l.smallWindows-1)

	success, err := l.script.Run(
		ctx, l.client, []string{resource}, l.window, l.limit, currentSmallWindow, startSmallWindow).Bool()
	if err != nil {
		return err
	}
	// 若到达窗口请求上限，请求失败
	if !success {
		return redisRateLimit.ErrAcquireFailed
	}
	return nil
}
