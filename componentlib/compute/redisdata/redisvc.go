package redisdata

import (
	"errors"

	"github.com/baishancloud/mallard/corelib/expvar"
	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/go-redis/redis"
)

var (
	log = zaplog.Zap("redis")

	queueCli *redis.Client
	cacheCli *redis.Client

	redisFailCount = expvar.NewDiff("redis.fail")
)

func init() {
	expvar.Register(redisFailCount)
}

var (
	// ErrQueueNil means queue db client is nil
	ErrQueueNil = errors.New("queue-client-nil")
	// ErrCacheNil means cache db client is nil
	ErrCacheNil = errors.New("cache-client-nil")
)

// SetClient sets clients
func SetClient(qCli, cCli *redis.Client) {
	queueCli = qCli
	cacheCli = cCli
}

// GetCacheCli gets cache redis db client
func GetCacheCli() *redis.Client {
	return cacheCli
}
