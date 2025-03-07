package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	Client *redis.Client
}

func NewRedRedisServie() *RedisService {
	return &RedisService{
		Client: redis.NewClient(
			&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			},
		),
	}
}

func (r *RedisService) PublishMessage(chatID string, message []byte) error {
	ctx := context.Background()

	return r.Client.Publish(ctx, chatID, message).Err()
}

func (r *RedisService) SubscribeMessages(chatID string) *redis.PubSub {
	ctx := context.Background()
	pubsub := r.Client.Subscribe(ctx, chatID)

	return pubsub
}
