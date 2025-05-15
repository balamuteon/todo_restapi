package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Options struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisClient(opt *Options) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     opt.Addr,
		Password: opt.Password,
		DB:       opt.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, err
	}
	return redisClient, nil
}
