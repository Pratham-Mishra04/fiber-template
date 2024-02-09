package cache

import (
	"context"
	"fmt"

	"github.com/Pratham-Mishra04/fiber-template/helpers"
	"github.com/Pratham-Mishra04/fiber-template/initializers"
	"github.com/redis/go-redis/v9"
)

var ctx = context.TODO()

func GetFromCache(key string) (string, error) {
	data, err := initializers.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("item not found in cache")
		}
		go helpers.LogServerError("Error Getting from cache", err, "")
		return "", fmt.Errorf("error getting from cache")
	}
	return data, nil
}

func SetToCache(key string, data []byte) error {
	if err := initializers.RedisClient.Set(ctx, key, data, initializers.CacheExpirationTime).Err(); err != nil {
		go helpers.LogServerError("Error Setting to cache", err, "")
		return fmt.Errorf("error setting to cache")
	}
	return nil
}

func RemoveFromCache(key string) error {
	err := initializers.RedisClient.Del(ctx, key).Err()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		go helpers.LogServerError("Error Removing from cache", err, "")
		return fmt.Errorf("error removing from cache")
	}
	return nil
}
