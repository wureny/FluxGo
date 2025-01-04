package algorithms

import (
	"context"
	"time"
)

// RateLimiter 定义了限流器的基本接口
type RateLimiter interface {
	// Allow 判断请求是否允许通过
	// key: 限流的唯一标识符(比如用户ID或IP)
	// 返回值: 是否允许请求通过，如果不允许还会返回需要等待的时间
	Allow(ctx context.Context, key string) (bool, time.Duration)

	// Close 清理资源
	Close() error
}

// Config 定义限流器的基本配置
type Config struct {
	// 时间窗口大小
	WindowSize time.Duration
	// 在窗口期内允许的最大请求数
	Limit int64
}
