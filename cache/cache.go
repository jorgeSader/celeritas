package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// Cache defines the interface for caching operations.
// It provides methods to check existence, retrieve, store, remove, and clear cache entries.
type Cache interface {
	// Has checks if a key exists in the cache.
	Has(string) (bool, error)
	// Get retrieves a value from the cache by key.
	Get(string) (interface{}, error)
	// Set stores a value in the cache with an optional expiration time.
	Set(string, interface{}, ...int) error
	// Forget removes a specific key from the cache.
	Forget(string) error
	// EmptyByMatch removes all cache entries matching a pattern.
	EmptyByMatch(string) error
	// Empty removes all cache entries with the cache prefix.
	Empty() error
}

// RedisCache represents a Redis-based cache implementation.
// It uses a connection pool and a prefix for key namespacing.
//
// The Prefix field should be unique per application to prevent key collisions
// when multiple applications share the same Redis instance.
type RedisCache struct {
	Conn   *redis.Pool // Redis connection pool
	Prefix string      // Namespace prefix for all keys (e.g., "app1")
}

// Entry is a map used to store cache data as key-value pairs.
// The "value" key holds the actual cached data, allowing for future metadata additions.
type Entry map[string]interface{}

// Has checks if a key exists in the Redis cache.
//
// The key is prefixed with the RedisCache Prefix (e.g., "prefix:key").
// It returns true if the key exists, false otherwise, and an error if the operation fails.
//
// Example:
//
//	cache := &RedisCache{Conn: pool, Prefix: "app1"}
//	exists, err := cache.Has("user")
//	// Checks for "app1:user" in Redis
func (c *RedisCache) Has(str string) (bool, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	conn := c.Conn.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return exists, nil
}

// encode serializes an Entry into a byte slice using gob encoding.
//
// It returns the encoded bytes and an error if encoding fails.
//
// Example:
//
//	entry := Entry{"value": "data"}
//	bytes, err := encode(entry)
func encode(item Entry) ([]byte, error) {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	if err := e.Encode(item); err != nil {
		return nil, fmt.Errorf("failed to encode entry: %w", err)
	}
	return b.Bytes(), nil
}

// decode deserializes a byte slice into an Entry using gob decoding.
//
// It returns the decoded Entry and an error if decoding fails.
//
// Example:
//
//	entry, err := decode(encodedBytes)
//	value := entry["value"]
func decode(data []byte) (Entry, error) {
	var item Entry
	b := bytes.NewBuffer(data)
	d := gob.NewDecoder(b)
	if err := d.Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}
	return item, nil
}

// Get retrieves a value from the Redis cache by key.
//
// The key is prefixed with the RedisCache Prefix (e.g., "prefix:key").
// It returns the cached value as an interface{} and an error if retrieval or decoding fails.
// Returns nil, nil if the key does not exist.
//
// Example:
//
//	cache := &RedisCache{Conn: pool, Prefix: "app1"}
//	value, err := cache.Get("user")
//	if err != nil { /* handle error */ }
//	if value != nil { /* use value */ }
func (c *RedisCache) Get(str string) (interface{}, error) {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	conn := c.Conn.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	data, err := redis.Bytes(conn.Do("GET", key))
	if err == redis.ErrNil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	decoded, err := decode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data for key %s: %w", key, err)
	}

	value, ok := decoded["value"]
	if !ok {
		return nil, fmt.Errorf("invalid cache format for key %s: missing 'value' field", key)
	}
	return value, nil
}

// Set stores a value in the Redis cache with an optional expiration time.
//
// The key is prefixed with the RedisCache Prefix (e.g., "prefix:key").
// The value is stored under the "value" key in an Entry map.
// The optional expires parameter specifies the TTL in seconds; if omitted, the key persists indefinitely.
// It returns an error if encoding or storage fails.
//
// Example:
//
//	cache := &RedisCache{Conn: pool, Prefix: "app1"}
//	err := cache.Set("user", "data", 3600) // Sets "app1:user" with 1-hour TTL
//	err := cache.Set("session", "token")   // Sets "app1:session" with no expiration
func (c *RedisCache) Set(str string, value interface{}, expires ...int) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	conn := c.Conn.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	entry := Entry{"value": value}
	encoded, err := encode(entry)
	if err != nil {
		return fmt.Errorf("failed to encode value for key %s: %w", key, err)
	}

	if len(expires) > 0 {
		_, err = conn.Do("SETEX", key, expires[0], encoded)
	} else {
		_, err = conn.Do("SET", key, encoded)
	}
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Forget removes a specific key from the Redis cache.
//
// The key is prefixed with the RedisCache Prefix (e.g., "prefix:key").
// It returns an error if the deletion fails.
//
// Example:
//
//	cache := &RedisCache{Conn: pool, Prefix: "app1"}
//	err := cache.Forget("user") // Deletes "app1:user"
func (c *RedisCache) Forget(str string) error {
	key := fmt.Sprintf("%s:%s", c.Prefix, str)
	conn := c.Conn.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	_, err := conn.Do("DEL", key)
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// EmptyByMatch removes all cache entries matching a pattern.
//
// The pattern is prefixed with the RedisCache Prefix (e.g., "prefix:pattern") and appended with ":*".
// It returns an error if key retrieval or deletion fails.
// If no keys match, it returns nil.
//
// Example:
//
//	cache := &RedisCache{Conn: pool, Prefix: "app1"}
//	err := cache.EmptyByMatch("user*") // Deletes all keys like "app1:user:*"
func (c *RedisCache) EmptyByMatch(pattern string) error {
	conn := c.Conn.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	matchPattern := fmt.Sprintf("%s:%s", c.Prefix, pattern)
	keys, err := c.getKeys(matchPattern)
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", matchPattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	_, err = conn.Do("DEL", args...)
	if err != nil {
		return fmt.Errorf("failed to delete %d keys for pattern %s: %w", len(keys), matchPattern, err)
	}
	return nil
}

// Empty removes all cache entries with the RedisCache Prefix.
//
// It matches all keys starting with "prefix:" (e.g., "prefix:*").
// It returns an error if key retrieval or deletion fails.
// If no keys match, it returns nil.
//
// Example:
//
//	cache := &RedisCache{Conn: pool, Prefix: "app1"}
//	err := cache.Empty() // Deletes all keys like "app1:*"
func (c *RedisCache) Empty() error {
	conn := c.Conn.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	pattern := fmt.Sprintf("%s:", c.Prefix)
	keys, err := c.getKeys(pattern)
	if err != nil {
		return fmt.Errorf("failed to get keys for prefix %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	_, err = conn.Do("DEL", args...)
	if err != nil {
		return fmt.Errorf("failed to delete %d keys for prefix %s: %w", len(keys), pattern, err)
	}
	return nil
}

// getKeys retrieves all Redis keys matching the given pattern using SCAN.
//
// The pattern is appended with ":*" to match all subkeys (e.g., "prefix:pattern:*").
// It returns the matched keys or an error if the operation fails.
// This method is used internally by Empty and EmptyByMatch.
//
// Example:
//
//	cache := &RedisCache{Conn: pool, Prefix: "app1"}
//	keys, err := cache.getKeys("user") // Gets all keys like "app1:user:*"
func (c *RedisCache) getKeys(pattern string) ([]string, error) {
	conn := c.Conn.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	var keys []string
	iter := 0
	matchPattern := pattern
	if !strings.HasSuffix(pattern, ":") {
		matchPattern = pattern + ":*"
	} else {
		matchPattern = pattern + "*"
	}

	for {
		scanResult, err := redis.Values(conn.Do("SCAN", iter, "MATCH", matchPattern, "COUNT", 1000))
		if err != nil {
			return keys, fmt.Errorf("scan failed for pattern %s: %w", matchPattern, err)
		}

		iter, err = redis.Int(scanResult[0], nil)
		if err != nil {
			return keys, fmt.Errorf("failed to parse scan iterator for pattern %s: %w", matchPattern, err)
		}

		matchedKeys, err := redis.Strings(scanResult[1], nil)
		if err != nil {
			return keys, fmt.Errorf("failed to parse matched keys for pattern %s: %w", matchPattern, err)
		}

		keys = append(keys, matchedKeys...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}
