package verified_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	redisV9 "github.com/wyy-go/wlimiter/verified/redis/v9"
	"github.com/wyy-go/wlimiter/verified/tests"
)

func TestReflux_Improve_Cover(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	defer mr.Close()
	tests.GenericTestReflux_Improve_Cover(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_RedisUnavailable(t *testing.T) {
	mr, err := miniredis.Run()
	require.Nil(t, err)
	addr := mr.Addr()
	mr.Close()
	tests.GenericTestReflux_RedisUnavailable(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: addr})),
	)
}

func TestReflux_One_Time(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestReflux_One_Time(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_In_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestReflux_In_Quota(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

func TestReflux_Over_Quota(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	defer mr.Close()

	tests.GenericTestReflux_Over_Quota(
		t,
		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
	)
}

// TODO: success in redis, but failed in miniredis
// func TestReflux_OneTime_Timeout(t *testing.T) {
// 	mr, err := miniredis.Run()
// 	assert.NoError(t, err)

// 	defer mr.Close()

// 	tests.GenericTestReflux_OneTime_Timeout(
// 		t,
// 		redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: mr.Addr()})),
// 		// redisV9.NewRedisStore(redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "123456", DB: 4})),
// 	)
// }
