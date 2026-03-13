package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

type Redis struct {
	Client *redis.Client
}

func NewRedis() *Redis {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:8080",
	})
	
	return &Redis{Client: client}
}