package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client — базовый клиент Redis
type Client struct {
	*redis.Client
}

// NewClient создает подключение к Redis
func NewClient(addr string, password string, db int) (*Client, error) {
	rclient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rclient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &Client{rclient}, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.Client.Close()
}

// Ping проверяет доступность
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// get Получаем значения из кеша
func (c *Client) get(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key).Result()
}

// set Сохраняем в кеш
func (c *Client) set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl).Err()
}

// setNX Сохраняем в кеш только если такого ключа еще нет
func (c *Client) setNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	return c.SetNX(ctx, key, value, ttl).Result()
}

// exists проверяет существует ли значение в кэше
func (c *Client) exists(ctx context.Context, key string) (bool, error) {
	n, err := c.Exists(ctx, key).Result()
	return n > 0, err
}

// delete Удаляем из кэша
func (c *Client) delete(ctx context.Context, keys ...string) error {
	return c.Del(ctx, keys...).Err()
}

// deleteByPrefix Удаляет все ключи, начинающиеся с указанного префикса
func (c *Client) deleteByPrefix(ctx context.Context, prefix string) error {
	var cursor uint64
	var keys []string
	var err error

	// Используем SCAN для итерации по всем ключам с заданным префиксом
	for {
		keys, cursor, err = c.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}

		// Если нашли ключи - удаляем их
		if len(keys) > 0 {
			if err := c.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		// Завершаем итерацию когда cursor вернет 0
		if cursor == 0 {
			break
		}
	}

	return nil
}
