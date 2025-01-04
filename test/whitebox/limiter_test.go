package whitebox

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wureny/FluxGo/internal/algorithms"
	"github.com/wureny/FluxGo/internal/algorithms/leakybucket"
	"github.com/wureny/FluxGo/internal/algorithms/slidinglog"
	"github.com/wureny/FluxGo/internal/algorithms/slidingwindow"
	"github.com/wureny/FluxGo/internal/algorithms/tokenbucket"
)

// 测试所有限流算法的基本功能
func TestRateLimiters(t *testing.T) {
	tests := []struct {
		name      string
		algorithm func(algorithms.Config) algorithms.RateLimiter
		limit     int64
		window    time.Duration
	}{
		{
			name: "SlidingLog",
			algorithm: func(c algorithms.Config) algorithms.RateLimiter {
				return slidinglog.NewLimiter(c)
			},
			limit:  10,
			window: time.Second,
		},
		{
			name: "SlidingWindow",
			algorithm: func(c algorithms.Config) algorithms.RateLimiter {
				return slidingwindow.NewLimiter(c)
			},
			limit:  10,
			window: time.Second,
		},
		{
			name: "LeakyBucket",
			algorithm: func(c algorithms.Config) algorithms.RateLimiter {
				return leakybucket.NewLimiter(c)
			},
			limit:  10,
			window: time.Second,
		},
		{
			name: "TokenBucket",
			algorithm: func(c algorithms.Config) algorithms.RateLimiter {
				return tokenbucket.NewLimiter(c)
			},
			limit:  10,
			window: time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := tt.algorithm(algorithms.Config{
				WindowSize: tt.window,
				Limit:      tt.limit,
			})
			defer limiter.Close()

			ctx := context.Background()
			key := "test-key"

			// 测试允许的请求
			for i := 0; i < int(tt.limit); i++ {
				allowed, wait := limiter.Allow(ctx, key)
				assert.True(t, allowed, "请求应该被允许")
				assert.Zero(t, wait, "不应该有等待时间")
			}

			// 测试超出限制
			allowed, wait := limiter.Allow(ctx, key)
			assert.False(t, allowed, "超出限制的请求应该被拒绝")
			assert.NotZero(t, wait, "应该有等待时间")

			// 等待时间窗口过期
			time.Sleep(tt.window)

			// 测试恢复后的请求
			allowed, wait = limiter.Allow(ctx, key)
			assert.True(t, allowed, "等待后请求应该被允许")
			assert.Zero(t, wait, "不应该有等待时间")
		})
	}
}
