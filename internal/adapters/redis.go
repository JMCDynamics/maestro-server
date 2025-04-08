package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	"github.com/go-redis/redis/v8"
)

type RedisCacheAdapter struct {
	client            *redis.Client
	expiredKeyChannel chan string
}

func NewRedisCacheAdapter(addr, password string, db int) interfaces.ICacheGateway {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	expiredKeyChannel := make(chan string)
	pubsub := rdb.PSubscribe(context.Background(), fmt.Sprintf("__keyevent@%d__:expired", db))

	go func() {
		for msg := range pubsub.Channel() {
			nodeID := msg.Payload
			expiredKeyChannel <- nodeID
		}
	}()

	return &RedisCacheAdapter{
		client:            rdb,
		expiredKeyChannel: expiredKeyChannel,
	}
}

func (r *RedisCacheAdapter) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCacheAdapter) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCacheAdapter) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCacheAdapter) ListenExpiredKeys() <-chan string {
	return r.expiredKeyChannel
}
