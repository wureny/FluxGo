package whitebox

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wureny/FluxGo/internal/algorithms"
	"github.com/wureny/FluxGo/internal/algorithms/slidinglog"
)

func TestLimiters(t *testing.T) {
	tests := []struct {
		name       string
		newLimiter func(algorithms.Config) algorithms.RateLimiter
		config     algorithms.Config
	}{
		{
			name: "SlidingLog",
			newLimiter: func(c algorithms.Config) algorithms.RateLimiter {
				return slidinglog.NewLimiter(c)
			},
			config: algorithms.Config{
				WindowSize: time.Second,
				Limit:      10,
			},
		},
		// ... 其他算法测试用例 ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := tt.newLimiter(tt.config)
			defer limiter.Close()

			ctx := context.Background()
			key := "test-key"

			// 测试正常请求
			for i := 0; i < int(tt.config.Limit); i++ {
				allowed, _ := limiter.Allow(ctx, key)
				assert.True(t, allowed, "请求应该被允许")
			}

			// 测试超出限制
			allowed, waitTime := limiter.Allow(ctx, key)
			assert.False(t, allowed, "超出限制的请求应该被拒绝")
			assert.True(t, waitTime > 0, "应该返回等待时间")

			// 测试等待后恢复
			time.Sleep(tt.config.WindowSize)
			allowed, _ = limiter.Allow(ctx, key)
			assert.True(t, allowed, "等待窗口期后请求应该被允许")
		})
	}
}
