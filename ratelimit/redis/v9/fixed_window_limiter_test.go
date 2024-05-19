package redis

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/wyy-go/wlimiter/ratelimit"
	"testing"
	"time"
)

func TestNewFixedWindowLimiter(t *testing.T) {
	type args struct {
		limit  int
		window time.Duration
	}
	tests := []struct {
		name string
		args args
		want ratelimit.RedisLimiter
	}{
		{
			name: "100_second",
			args: args{
				limit:  100,
				window: time.Second,
			},
			want: nil,
		},
	}

	mr, err := miniredis.Run()
	assert.Nil(t, err)

	defer mr.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := redis.NewClient(&redis.Options{
				Addr: mr.Addr(),
			})
			l, _ := NewFixedWindowLimiter(client, tt.args.limit, tt.args.window)
			successCount := 0
			for i := 0; i < tt.args.limit*2; i++ {
				if l.TryAcquire(context.Background(), "test") == nil {
					successCount++
				}
			}
			if successCount != tt.args.limit {
				t.Errorf("NewFixedWindowLimiter() = %v, want %v", successCount, tt.args.limit)
			}
			time.Sleep(time.Second)
			successCount = 0
			for i := 0; i < tt.args.limit*2; i++ {
				if l.TryAcquire(context.Background(), "test") == nil {
					successCount++
				}
			}
			if successCount != tt.args.limit {
				t.Errorf("NewFixedWindowLimiter() = %v, want %v", successCount, tt.args.limit)
			}
		})
	}
}
