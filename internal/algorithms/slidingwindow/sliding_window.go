package slidingwindow

import (
	"context"
	"sync"
	"time"

	"github.com/wureny/FluxGo/internal/algorithms"
)

// 窗口计数记录
type windowCount struct {
	count     int64     // 当前窗口的请求计数
	timestamp time.Time // 窗口的起始时间
}

// SlidingWindowLimiter 实现基于滑动窗口计数的限流器
type SlidingWindowLimiter struct {
	mu sync.RWMutex
	// key -> 窗口计数的映射
	windows map[string]windowCount
	// 配置信息
	config algorithms.Config
}

// NewLimiter 创建一个新的滑动窗口计数限流器
func NewLimiter(config algorithms.Config) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windows: make(map[string]windowCount),
		config:  config,
	}
}

// Allow 实现RateLimiter接口
func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	window, exists := l.windows[key]

	// 如果窗口不存在或已过期，创建新窗口
	if !exists || now.Sub(window.timestamp) >= l.config.WindowSize {
		l.windows[key] = windowCount{
			count:     1,
			timestamp: now,
		}
		return true, 0
	}

	// 计算当前请求数量是否超过限制
	if window.count >= l.config.Limit {
		waitDuration := window.timestamp.Add(l.config.WindowSize).Sub(now)
		return false, waitDuration
	}

	// 更新计数
	window.count++
	l.windows[key] = window
	return true, 0
}

// Close 实现RateLimiter接口
func (l *SlidingWindowLimiter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.windows = make(map[string]windowCount)
	return nil
}
