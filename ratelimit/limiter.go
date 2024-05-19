package ratelimit

import "context"

type Limiter interface {
	TryAcquire() error
}

type RedisLimiter interface {
	TryAcquire(ctx context.Context, source string) error
}
