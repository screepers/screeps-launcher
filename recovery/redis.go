package recovery

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
)

type redisData map[string]redisValue

type redisValue struct {
	TTL   time.Duration
	Value string
}

func (r *Recovery) getRedisClient() *redis.Client {
	host := "localhost"
	port := "6379"
	if v, ok := r.config.Env.Shared["REDIS_HOST"]; ok {
		host = v
	}
	if v, ok := r.config.Env.Shared["REDIS_PORT"]; ok {
		port = v
	}
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "",
		DB:       0,
	})
	return client
}

func (r *Recovery) redisBackup() (redisData, error) {
	client := r.getRedisClient()
	keys, err := client.Keys("*").Result()
	if err != nil {
		return nil, err
	}
	col := redisData{}
	for _, key := range keys {
		value, err := client.Dump(key).Result()
		if err != nil {
			return nil, err
		}
		ttl := client.TTL(key).Val()
		col[key] = redisValue{
			TTL:   ttl,
			Value: value,
		}
	}
	return col, nil
}

func (r *Recovery) redisRestore(data redisData) error {
	client := r.getRedisClient()
	if err := client.FlushAll().Err(); err != nil {
		return err
	}
	for k, v := range data {
		if err := client.RestoreReplace(k, v.TTL, v.Value).Err(); err != nil {
			return err
		}
	}
	return nil
}
