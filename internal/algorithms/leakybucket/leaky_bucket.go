package leakybucket

import (
	"context"
	"sync"
	"time"

	"github.com/wureny/FluxGo/internal/algorithms"
)

// 漏桶状态
type bucket struct {
	water        float64   // 当前水量
	lastLeakTime time.Time // 上次漏水时间
}

// LeakyBucketLimiter 实现基于漏桶算法的限流器
type LeakyBucketLimiter struct {
	mu sync.RWMutex
	// key -> 漏桶的映射
	buckets map[string]bucket
	// 配置信息
	config algorithms.Config
	// 漏水速率（每秒）
	rate float64
	// 桶容量
	capacity float64
}

// NewLimiter 创建一个新的漏桶限流器
func NewLimiter(config algorithms.Config) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		buckets:  make(map[string]bucket),
		config:   config,
		rate:     float64(config.Limit) / config.WindowSize.Seconds(),
		capacity: float64(config.Limit),
	}
}

// Allow 实现RateLimiter接口
func (l *LeakyBucketLimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.buckets[key]
	if !exists {
		// 新建漏桶
		l.buckets[key] = bucket{
			water:        1, // 初始水量为1（当前请求）
			lastLeakTime: now,
		}
		return true, 0
	}

	// 计算从上次漏水到现在流出的水量
	elapsed := now.Sub(b.lastLeakTime).Seconds()
	leakedWater := elapsed * l.rate
	currentWater := max(0, b.water-leakedWater)

	// 如果加入当前请求后会溢出，则拒绝请求
	if currentWater+1 > l.capacity {
		waitTime := time.Duration((currentWater + 1 - l.capacity) / l.rate * float64(time.Second))
		return false, waitTime
	}

	// 更新水量和时间
	b.water = currentWater + 1
	b.lastLeakTime = now
	l.buckets[key] = b

	return true, 0
}

// Close 实现RateLimiter接口
func (l *LeakyBucketLimiter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = make(map[string]bucket)
	return nil
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
