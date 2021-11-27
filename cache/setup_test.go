package cache

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"os"
	"testing"
	"time"
)

var testRedisCache RedisCache

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	pool := redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
		MaxIdle:     50,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
	}

	testRedisCache.Conn = &pool
	testRedisCache.Prefix = "test-bendis"

	defer testRedisCache.Conn.Close()

	os.Exit(m.Run())
}
