package models

import (
	"errors"

	"github.com/go-redis/redis"
)

func RedisKeyCreate(r *redis.Client, key string, val redis.Z) error {
	if res := r.ZAddNX(key, val); res.Val() == 0 {
		return errors.New("Entry was not inserted")
	}
	return nil
}

func RedisKeyExists(r *redis.Client, key, val string) bool {
	zScore := r.ZScore(key, val)

	if _, scoreErr := zScore.Result(); scoreErr == nil {
		return true
	}
	return false
}
