package auth

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisStore struct {
	r *redis.Client
}

func (rs RedisStore) Set(runId string, jwt string) error {
	// For now, we'll set tokens to a valid span of 5 hours
	return rs.r.Set(context.TODO(), runId, jwt, time.Hour*5).Err()
}

func (rs RedisStore) Check(runId string, jwt string) (bool, error) {
	val, err := rs.r.Get(context.TODO(), runId).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if val != jwt {
		return false, nil
	}
	return true, nil
}

// For now, we won't delete tokens, just leverage the expiration times
func (rs RedisStore) Delete(runId string) error {
	return nil
}
