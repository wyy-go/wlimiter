package redis

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/wyy-go/wlimiter/ratelimit"
	"testing"
	"time"
)

func TestNewLeakyBucketLimiter(t *testing.T) {
	type args struct {
		peakLevel       int
		currentVelocity int
	}
	tests := []struct {
		name    string
		args    args
		want    ratelimit.RedisLimiter
		wantErr bool
	}{
		{
			name: "60",
			args: args{
				peakLevel:       60,
				currentVelocity: 10,
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
			l := NewLeakyBucketLimiter(client, tt.args.peakLevel, tt.args.currentVelocity)
			successCount := 0
			for i := 0; i < tt.args.peakLevel*2; i++ {
				if l.TryAcquire(context.Background(), "test") == nil {
					successCount++
				}
			}
			if successCount != tt.args.peakLevel {
				t.Errorf("NewLeakyBucketLimiter() got = %v, want %v", successCount, tt.args.peakLevel)
				return
			}

			time.Sleep(time.Second * time.Duration(tt.args.peakLevel/tt.args.currentVelocity) / 2)
			successCount = 0
			for i := 0; i < tt.args.peakLevel; i++ {
				if l.TryAcquire(context.Background(), "test") == nil {
					successCount++
				}
			}
			if successCount != tt.args.peakLevel/2 {
				t.Errorf("NewLeakyBucketLimiter() got = %v, want %v", successCount, tt.args.peakLevel/2)
				return
			}
		})
	}
}
