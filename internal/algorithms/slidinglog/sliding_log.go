package slidinglog

import (
	"context"
	"sync"
	"time"

	"github.com/wureny/FluxGo/internal/algorithms"
)

// 请求记录
type requestLog struct {
	timestamp time.Time
}

// SlidingLogLimiter 实现基于滑动窗口日志的限流器
type SlidingLogLimiter struct {
	mu sync.RWMutex
	// key -> 请求日志切片的映射
	logs map[string][]requestLog
	// 配置信息
	config algorithms.Config
}

// NewLimiter 创建一个新的滑动窗口日志限流器
func NewLimiter(config algorithms.Config) *SlidingLogLimiter {
	return &SlidingLogLimiter{
		logs:   make(map[string][]requestLog),
		config: config,
	}
}

// Allow 实现RateLimiter接口
func (l *SlidingLogLimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.config.WindowSize)

	// 获取该key的请求日志
	logs := l.logs[key]

	// 清理过期的日志
	validLogs := make([]requestLog, 0)
	for _, log := range logs {
		if log.timestamp.After(windowStart) {
			validLogs = append(validLogs, log)
		}
	}

	// 如果请求数量未达到限制，允许请求
	if int64(len(validLogs)) < l.config.Limit {
		l.logs[key] = append(validLogs, requestLog{timestamp: now})
		return true, 0
	}

	// 计算需要等待的时间
	waitDuration := validLogs[0].timestamp.Add(l.config.WindowSize).Sub(now)
	return false, waitDuration
}

// Close 实现RateLimiter接口
func (l *SlidingLogLimiter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 清理所有数据
	l.logs = make(map[string][]requestLog)
	return nil
}
