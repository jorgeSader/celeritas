package cache

import (
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
)

var testRedisCache RedisCache
var testRedisServer *miniredis.Miniredis // Add server as global var

func TestMain(m *testing.M) {
	// Start miniredis server
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	testRedisServer = s // Store the server instance
	defer s.Close()

	// Configure connection pool
	pool := redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", s.Addr())
		},
	}

	// Initialize test cache
	testRedisCache = RedisCache{
		Conn:   &pool,
		Prefix: "test-devify",
	}

	// Ensure pool is closed after tests
	defer func() {
		if err := pool.Close(); err != nil {
			println("Failed to close Redis pool:", err.Error())
		}
	}()

	// Run tests and exit
	os.Exit(m.Run())
}

func resetCache() error {
	return testRedisCache.Empty()
}
