package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/wyy-go/wlimiter/ratelimit"
	redisRateLimit "github.com/wyy-go/wlimiter/ratelimit/redis"
	"sort"
	"time"
)

// ViolationStrategyError 违背策略错误
type ViolationStrategyError struct {
	Limit  int           // 窗口请求上限
	Window time.Duration // 窗口时间大小
}

func (e *ViolationStrategyError) Error() string {
	return fmt.Sprintf("violation strategy that limit = %d and window = %d", e.Limit, e.Window)
}

// SlidingLogLimiterStrategy 滑动日志限流器的策略
type SlidingLogLimiterStrategy struct {
	limit        int   // 窗口请求上限
	window       int64 // 窗口时间大小
	smallWindows int64 // 小窗口数量
}

func NewSlidingLogLimiterStrategy(limit int, window time.Duration) *SlidingLogLimiterStrategy {
	return &SlidingLogLimiterStrategy{
		limit:  limit,
		window: int64(window),
	}
}

// SlidingLogLimiter 滑动日志限流器
type slidingLogLimiter struct {
	strategies  []*SlidingLogLimiterStrategy // 滑动日志限流器策略列表
	smallWindow int64                        // 小窗口时间大小
	client      *redis.Client                // Redis客户端
	script      *redis.Script                // TryAcquire脚本
}

func NewSlidingLogLimiter(client *redis.Client, smallWindow time.Duration, strategies ...*SlidingLogLimiterStrategy) (
	ratelimit.RedisLimiter, error) {
	// 复制策略避免被修改
	strategies = append(make([]*SlidingLogLimiterStrategy, 0, len(strategies)), strategies...)

	// 不能不设置策略
	if len(strategies) == 0 {
		return nil, errors.New("must be set strategies")
	}

	// redis过期时间精度最大到毫秒，因此窗口必须能被毫秒整除
	if smallWindow%time.Millisecond != 0 {
		return nil, errors.New("the window uint must not be less than millisecond")
	}
	smallWindow = smallWindow / time.Millisecond
	for _, strategy := range strategies {
		if strategy.window%int64(time.Millisecond) != 0 {
			return nil, errors.New("the window uint must not be less than millisecond")
		}
		strategy.window = strategy.window / int64(time.Millisecond)
	}

	// 排序策略，窗口时间大的排前面，相同窗口上限大的排前面
	sort.Slice(strategies, func(i, j int) bool {
		a, b := strategies[i], strategies[j]
		if a.window == b.window {
			return a.limit > b.limit
		}
		return a.window > b.window
	})

	for i, strategy := range strategies {
		// 随着窗口时间变小，窗口上限也应该变小
		if i > 0 {
			if strategy.limit >= strategies[i-1].limit {
				return nil, errors.New("the smaller window should be the smaller limit")
			}
		}
		// 窗口时间必须能够被小窗口时间整除
		if strategy.window%int64(smallWindow) != 0 {
			return nil, errors.New("window cannot be split by integers")
		}
		strategy.smallWindows = strategy.window / int64(smallWindow)
	}

	return &slidingLogLimiter{
		strategies:  strategies,
		smallWindow: int64(smallWindow),
		client:      client,
		script:      redis.NewScript(redisRateLimit.SlidingLogLimitScript),
	}, nil
}

func (l *slidingLogLimiter) TryAcquire(ctx context.Context, resource string) error {
	// 获取当前小窗口值
	currentSmallWindow := time.Now().UnixMilli() / l.smallWindow * l.smallWindow
	args := make([]interface{}, len(l.strategies)*2+2)
	args[0] = currentSmallWindow
	args[1] = l.strategies[0].window
	// 获取每个策略的起始小窗口值
	for i, strategy := range l.strategies {
		args[i*2+2] = currentSmallWindow - l.smallWindow*(strategy.smallWindows-1)
		args[i*2+3] = strategy.limit
	}

	index, err := l.script.Run(
		ctx, l.client, []string{resource}, args...).Int()
	if err != nil {
		return err
	}
	// 若到达窗口请求上限，请求失败
	if index != -1 {
		return &ViolationStrategyError{
			Limit:  l.strategies[index].limit,
			Window: time.Duration(l.strategies[index].window),
		}
	}
	return nil
}
