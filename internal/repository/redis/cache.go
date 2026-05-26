package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Cache предоставляет операции кэширования
type Cache struct {
	client *Client
	ttl    time.Duration
}

// NewCache создает новый экземпляр кэша
func NewCache(client *Client, ttl time.Duration) *Cache {
	return &Cache{
		client: client,
		ttl:    ttl,
	}
}

// Get получает значение из кэша
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.get(ctx, key)
}

// GetJSON получает и десериализует JSON из кэша
func (c *Cache) GetJSON(ctx context.Context, key string, dest any) error {
	data, err := c.client.get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Set сохраняет значение в кэш с TTL по умолчанию
func (c *Cache) Set(ctx context.Context, key string, value any) error {
	return c.client.set(ctx, key, value, c.ttl)
}

// SetJSON сериализует в JSON и сохраняет
func (c *Cache) SetJSON(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}
	return c.client.set(ctx, key, data, c.ttl)
}

// SetWithTTL сохраняет с указанным TTL
func (c *Cache) SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.client.set(ctx, key, value, ttl)
}

// Exists проверяет наличие ключа
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	return c.client.exists(ctx, key)
}

// Delete удаляет ключ
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.delete(ctx, key)
}

// DeleteByPrefix удаляет все ключи с префиксом
func (c *Cache) DeleteByPrefix(ctx context.Context, prefix string) error {
	return c.client.deleteByPrefix(ctx, prefix)
}
