package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/redis/go-redis/v9"
)

type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(addr string) *RedisSessionStore {
	return &RedisSessionStore{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (s *RedisSessionStore) Save(ctx context.Context, key string, data *webauthn.SessionData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, b, 5*time.Minute).Err()
}

func (s *RedisSessionStore) Load(ctx context.Context, key string) (*webauthn.SessionData, error) {
	b, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found or expired")
	} else if err != nil {
		return nil, err
	}

	var data webauthn.SessionData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, key string) {
	s.client.Del(ctx, key)
}
