package service

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"
	"gopkg.in/redis.v3"
)

type Redis struct {
	*redis.Client
}

//  Delete all Redis keys with a given prefix wildcard, e.g. "data:*"
func (r *Redis) DeleteKeysPrefix(prefix string) (interface{}, error) {
	script := "return redis.call('del', unpack(redis.call('keys', ARGV[1])))"
	return r.Eval(script, []string{}, []string{prefix}).Result()
}

func (r *Redis) KeyCount(pattern string) int {
	var cursor int64
	var n int

	for {
		var keys []string
		var err error
		cursor, keys, err = r.Scan(cursor, pattern, 10).Result()
		if err != nil {
			n = -1
			break
		} else {
			n += len(keys)
			if cursor == 0 {
				break
			}
		}
	}

	return n
}

// Write a protocol buffer to cache with the provided key and expiry
func (r *Redis) SetCachedProtobuf(key string, obj proto.Message, expiry time.Duration) error {
	msg, err := proto.Marshal(obj)
	if err != nil {
		return err
	}
	r.Set(key, msg, expiry)

	return nil
}

// Add a protocol buffer to the Redis set at the given key
func (r *Redis) SAddCachedProtobuf(key string, obj proto.Message) error {
	msg, err := proto.Marshal(obj)
	if err != nil {
		return err
	}
	r.SAdd(key, string(msg))

	return nil
}

// Read a protocol buffer from the cache
func (r *Redis) GetCachedProtobuf(key string, obj proto.Message) ([]byte, error) {
	value, err := r.Get(key).Result()
	if err != nil {
		return nil, err
	}

	bytesArray := []byte(value)
	err = proto.Unmarshal(bytesArray, obj)

	return bytesArray, err
}

// Write a HTTP response with content from the cached object with the given key and protocol buffer type
func (r *Redis) WriteProtobufKey(w http.ResponseWriter, key string, obj proto.Message, writeJson bool) error {
	bytes, err := r.GetCachedProtobuf(key, obj)
	if err == redis.Nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		if writeJson {
			WriteJsonResponse(w, obj)
		} else {
			WriteBytes(w, bytes)
		}
	}

	return err
}

// j
func (r *Redis) WriteCacheProtobufMessage(w http.ResponseWriter, obj proto.Message, cacheKey string, expiry time.Duration, useJson bool) {
	msg, _ := proto.Marshal(obj)
	r.Set(cacheKey, msg, expiry)

	if useJson {
		WriteJsonResponse(w, obj)
	} else {
		WriteBytes(w, []byte(msg))
	}
}

func (r *Redis) GetProtobufKey(key string, obj proto.Message) error {
	value, err := r.Get(key).Result()
	if err != nil {
		return err
	}

	err = proto.Unmarshal([]byte(value), obj)
	if err != nil {
		return err
	}

	return nil
}

// Save a json object to Redis
func (r *Redis) CacheJson(key string, value interface{}, expiry time.Duration) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		Log.Error.Printf("Error marshalling to cache: %v", err)
	} else {
		r.Set(key, string(jsonData[:]), 0)
		if expiry > 0 {
			r.Expire(key, expiry)
		}
	}
}

// Read a protocol buffer from the cache
func (r *Redis) ReadJson(key string, result interface{}) error {
	data, err := r.Get(key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

type CacheWriter struct {
	key    string
	redis  *redis.Client
	expiry time.Duration
}

func (cw CacheWriter) Write(p []byte) (n int, err error) {
	//  append to key
	str, err := cw.redis.Get(cw.key).Result()
	if err != nil {
		return 0, err
	}

	str += string(p[:])
	cw.redis.Set(cw.key, str, cw.expiry)

	return len(p), nil
}

//  Returns a Writer that will output to a Redis key
func (r *Redis) CacheKeyWriter(key string, expiry time.Duration) io.Writer {
	cw := CacheWriter{key: key, redis: r.Client, expiry: expiry}
	r.Set(key, "", expiry)
	return cw
}

func (r *Redis) RecordTimeKey(key string) {
	r.Set(key, PerthNow().String(), 0)
}

// Remove all Redis cache entries matching a glob pattern
func (r *Redis) ClearRedisKeys(glob string) error {

	_, err := r.Eval("return redis.call('del', unpack(redis.call('keys', ARGV[1])))",
		[]string{},
		[]string{glob}).Result()

	return err
}

func (r *Redis) StatRecordIncr(key string, score float64, member string) error {
	_, err := r.ZIncrBy(key, score, member).Result()
	return err
}

func (r *Redis) StatRevRange(key string) ([]string, error) {
	r.ZRemRangeByRank(key, 0, -100)
	return r.ZRevRange(key, 0, 100).Result()
}

func (r *Redis) StatRecordHourValue(key string, value string) {
	now := PerthNow()

	cacheKey := fmt.Sprintf("%v:%v", key, now.Format("20060102"))
	hourKey := now.Format("15")

	r.HSet(cacheKey, hourKey, value)
	r.Expire(cacheKey, 24*time.Hour)
}

func ReadProtobuf(reader io.Reader, message proto.Message) error {
	body, err := ioutil.ReadAll(reader)
	if err == nil {
		err = proto.Unmarshal(body, message)
	}

	return err
}

func WriteProtobufMessage(w http.ResponseWriter, obj proto.Message, useJson bool) {
	if useJson {
		WriteJsonResponse(w, obj)
	} else {
		msg, _ := proto.Marshal(obj)
		WriteBytes(w, []byte(msg))
	}
}
