package redis

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/wyy-go/wlimiter/ratelimit"
	"testing"
	"time"
)

func TestNewSlidingLogLimiter(t *testing.T) {
	type args struct {
		smallWindow time.Duration
		strategies  []*SlidingLogLimiterStrategy
	}
	tests := []struct {
		name    string
		args    args
		want    ratelimit.RedisLimiter
		wantErr bool
	}{
		{
			name: "60_5seconds",
			args: args{
				smallWindow: time.Second,
				strategies: []*SlidingLogLimiterStrategy{
					NewSlidingLogLimiterStrategy(10, time.Minute),
					NewSlidingLogLimiterStrategy(100, time.Hour),
				},
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
			NewSlidingLogLimiter(client, tt.args.smallWindow, tt.args.strategies...)
		})
	}
}
