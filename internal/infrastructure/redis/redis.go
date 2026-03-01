package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client wraps the go-redis connection exposing robust ping controls.
type Client struct {
	Client *goredis.Client
}

// NewRedisClient leverages graceful degradation. It attempts a ping; if offline, it returns nil instead of crashing, allowing API endpoints to fall-back reliably.
func NewRedisClient(ctx context.Context, url, host, port, password string) *Client {
	var opts *goredis.Options
	var addr string

	if url != "" {
		var err error
		opts, err = goredis.ParseURL(url)
		if err != nil {
			slog.Error("Failed to parse REDIS_URL", "error", err)
			return nil
		}
		addr = opts.Addr
	} else {
		addr = fmt.Sprintf("%s:%s", host, port)
		opts = &goredis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
			PoolSize: 10,
		}
	}

	rdb := goredis.NewClient(opts)

	// Ping mapping ensuring short execution bursts validating liveliness.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("Redis cluster is completely unreachable. Proceeding with Graceful Degradation (cache bypassed)", "error", err, "address", addr)
		// We deliberately return 'nil' here because the system must survive without caching.
		return nil
	}

	slog.Info("Successfully established connection to Redis Cluster", "address", addr)
	return &Client{Client: rdb}
}
