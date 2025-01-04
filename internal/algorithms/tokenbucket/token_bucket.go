package tokenbucket

import (
	"context"
	"sync"
	"time"

	"github.com/wureny/FluxGo/internal/algorithms"
)

// 令牌桶状态
type bucket struct {
	tokens     float64   // 当前令牌数
	lastRefill time.Time // 上次补充令牌的时间
}

// TokenBucketLimiter 实现基于令牌桶算法的限流器
type TokenBucketLimiter struct {
	mu sync.RWMutex
	// key -> 令牌桶的映射
	buckets map[string]bucket
	// 配置信息
	config algorithms.Config
	// 令牌生成速率（每秒）
	rate float64
	// 桶容量
	capacity float64
}

// NewLimiter 创建一个新的令牌桶限流器
func NewLimiter(config algorithms.Config) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets:  make(map[string]bucket),
		config:   config,
		rate:     float64(config.Limit) / config.WindowSize.Seconds(),
		capacity: float64(config.Limit),
	}
}

// Allow 实现RateLimiter接口
func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.buckets[key]
	if !exists {
		// 新建令牌桶，初始容量为满
		l.buckets[key] = bucket{
			tokens:     l.capacity - 1, // 减1是因为当前请求会消耗一个令牌
			lastRefill: now,
		}
		return true, 0
	}

	// 计算需要补充的令牌数
	elapsed := now.Sub(b.lastRefill).Seconds()
	newTokens := elapsed * l.rate
	b.tokens = min(l.capacity, b.tokens+newTokens)
	b.lastRefill = now

	// 如果没有令牌，拒绝请求
	if b.tokens < 1 {
		waitTime := time.Duration((1 - b.tokens) / l.rate * float64(time.Second))
		return false, waitTime
	}

	// 消耗令牌
	b.tokens--
	l.buckets[key] = b
	return true, 0
}

// Close 实现RateLimiter接口
func (l *TokenBucketLimiter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = make(map[string]bucket)
	return nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
