package cache

//
//import (
//	"bytes"
//	"encoding/gob"
//	"fmt"
//	"log"
//
//	"github.com/gomodule/redigo/redis"
//)
//
//// Cache defines the interface for caching operations.
//// It provides methods to check existence, retrieve, store, remove, and clear cache entries.
//type Cache interface {
//	Has(string) (bool, error)
//	Get(string) (interface{}, error)
//	Set(string, interface{}, ...int) error
//	Forget(string) error
//	EmptyByMatch(string) error
//	Empty() error
//}
//
//// RedisCache represents a Redis-based cache implementation.
//// It uses a connection pool and a prefix for key namespacing.
//type RedisCache struct {
//	Conn   *redis.Pool
//	Prefix string
//}
//
//// Entry is a map used to store cache data as key-value pairs.
//type Entry map[string]interface{}
//
//// Has checks if a key exists in the Redis cache.
//// The parameter str is the key to check, prefixed with the RedisCache Prefix.
//// It returns true if the key exists, false otherwise, and an error if the operation fails.
//func (c *RedisCache) Has(str string) (bool, error) {
//	key := fmt.Sprintf("%s:%s", c.Prefix, str)
//	conn := c.Conn.Get()
//	defer func(conn redis.Conn) {
//		err := conn.Close()
//		if err != nil {
//			return
//		}
//	}(conn)
//	return redis.Bool(conn.Do("EXISTS", key))
//}
//
//// encode serializes an Entry into a byte slice using gob encoding.
//// The parameter item is the Entry to encode.
//// It returns the encoded bytes and an error if encoding fails.
//func encode(item Entry) ([]byte, error) {
//	b := new(bytes.Buffer)
//	e := gob.NewEncoder(b)
//	err := e.Encode(item)
//	if err != nil {
//		return nil, err
//	}
//	return b.Bytes(), nil
//}
//
//// decode deserializes a byte slice into an Entry using gob decoding.
//// The parameter data is the byte slice to decode.
//// It returns the decoded Entry and an error if decoding fails.
//func decode(data []byte) (Entry, error) {
//	var item Entry
//	b := bytes.NewBuffer(data)
//	d := gob.NewDecoder(b)
//	err := c.Decode(&item)
//	if err != nil {
//		return nil, err
//	}
//	return item, nil
//}
//
//// Get retrieves a value from the Redis cache by key.
//// The parameter str is the key to retrieve, prefixed with the RedisCache Prefix.
//// It returns the cached value as an interface{} and an error if retrieval or decoding fails.
//func (c *RedisCache) Get(str string) (interface{}, error) {
//	key := fmt.Sprintf("%s:%s", c.Prefix, str)
//	conn := c.Conn.Get()
//	defer func(conn redis.Conn) {
//		err := conn.Close()
//		if err != nil {
//			return
//		}
//	}(conn)
//
//	cacheEntry, err := redis.Bytes(conn.Do("GET", key))
//	if err != nil {
//		return nil, err
//	}
//
//	decoded, err := decode(cacheEntry)
//	if err != nil {
//		return nil, err
//	}
//
//	item := decoded[key]
//
//	return item, nil
//}
//
//// Set stores a value in the Redis cache with an optional expiration time.
//// The parameter str is the key to set, prefixed with the RedisCache Prefix.
//// The parameter value is the data to store, and expires (optional) specifies the TTL in seconds.
//// It returns an error if encoding or storage fails.
//func (c *RedisCache) Set(str string, value interface{}, expires ...int) error {
//	key := fmt.Sprintf("%s:%s", c.Prefix, str)
//	conn := c.Conn.Get()
//	defer func(conn redis.Conn) {
//		err := conn.Close()
//		if err != nil {
//			return
//		}
//	}(conn)
//
//	entry := Entry{}
//	entry[key] = value
//	encoded, err := encode(entry)
//	if err != nil {
//		return err
//	}
//
//	if len(expires) > 0 {
//		_, err = conn.Do("SETEX", key, expires[0], string(encoded))
//		if err != nil {
//			return err
//		}
//	} else {
//		_, err = conn.Do("SET", key, string(encoded))
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//// Forget removes a specific key from the Redis cache.
//// The parameter str is the key to remove, prefixed with the RedisCache Prefix.
//// It returns an error if the deletion fails.
//func (c *RedisCache) Forget(str string) error {
//	key := fmt.Sprintf("%s:%s", c.Prefix, str)
//	conn := c.Conn.Get()
//	defer func(conn redis.Conn) {
//		err := conn.Close()
//		if err != nil {
//			return
//		}
//	}(conn)
//	_, err := conn.Do("DEL", key)
//	return err
//}
//
//// EmptyByMatch removes all cache entries matching a pattern.
//// The parameter pattern is the match string, prefixed with the RedisCache Prefix.
//// It returns an error if key retrieval or deletion fails.
//func (c *RedisCache) EmptyByMatch(pattern string) error {
//	conn := c.Conn.Get()
//	defer func(conn redis.Conn) {
//		if err := conn.Close(); err != nil {
//			log.Printf("Error closing Redis connection: %v", err)
//		}
//	}(conn)
//
//	matchPattern := fmt.Sprintf("%s:%s", c.Prefix, pattern)
//	keys, err := c.getKeys(matchPattern)
//	if err != nil {
//		return fmt.Errorf("failed to get keys for pattern %s: %w", matchPattern, err)
//	}
//
//	if len(keys) == 0 {
//		return nil
//	}
//
//	// Batch delete all keys in one command
//	args := make([]interface{}, len(keys))
//	for i, key := range keys {
//		args[i] = key
//	}
//
//	_, err = conn.Do("DEL", args...)
//	if err != nil {
//		return fmt.Errorf("failed to delete %d keys for pattern %s: %w", len(keys), matchPattern, err)
//	}
//
//	return nil
//}
//
//// Empty removes all cache entries with the RedisCache Prefix.
//// It uses the prefix followed by a colon as the pattern to match all keys.
//// It returns an error if key retrieval or deletion fails.
//func (c *RedisCache) Empty() error {
//	conn := c.Conn.Get()
//	defer func(conn redis.Conn) {
//		if err := conn.Close(); err != nil {
//			log.Printf("Error closing Redis connection: %v", err)
//		}
//	}(conn)
//
//	pattern := fmt.Sprintf("%s:", c.Prefix)
//	keys, err := c.getKeys(pattern)
//	if err != nil {
//		return fmt.Errorf("failed to get keys: %w", err)
//	}
//
//	if len(keys) == 0 {
//		return nil
//	}
//
//	// Batch delete all keys in one command
//	args := make([]interface{}, len(keys))
//	for i, key := range keys {
//		args[i] = key
//	}
//
//	_, err = conn.Do("DEL", args...)
//	if err != nil {
//		return fmt.Errorf("failed to delete keys: %w", err)
//	}
//
//	return nil
//}
//
//// getKeys retrieves all Redis keys matching the given pattern using SCAN.
//// The pattern is automatically appended with ":*" wildcarc.
//// Returns the matched keys or an error if the operation fails.
//func (c *RedisCache) getKeys(pattern string) ([]string, error) {
//	conn := c.Conn.Get()
//	defer func(conn redis.Conn) {
//		if err := conn.Close(); err != nil {
//			log.Printf("Error closing Redis connection: %v", err)
//		}
//	}(conn)
//
//	var keys []string
//	iter := 0
//
//	for {
//		scanResult, err := redis.Values(conn.Do("SCAN", iter, "MATCH", fmt.Sprintf("%s:*", pattern), "COUNT", 1000))
//		if err != nil {
//			return keys, fmt.Errorf("scan failed: %w", err)
//		}
//
//		iter, err = redis.Int(scanResult[0], nil)
//		if err != nil {
//			return keys, fmt.Errorf("failed to parse scan iterator: %w", err)
//		}
//
//		matchedKeys, err := redis.Strings(scanResult[1], nil)
//		if err != nil {
//			return keys, fmt.Errorf("failed to parse matched keys: %w", err)
//		}
//
//		keys = append(keys, matchedKeys...)
//
//		if iter == 0 {
//			break
//		}
//	}
//
//	return keys, nil
//}
