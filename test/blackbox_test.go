package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wureny/FluxGo/internal/gateway"
	"github.com/wureny/FluxGo/internal/limiter"
	"github.com/wureny/FluxGo/pkg/client"
)

func TestEndToEnd(t *testing.T) {
	// 创建测试API服务器
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "ok", "path": "%s"}`, r.URL.Path)
	}))
	defer apiServer.Close()

	// 创建网关
	gw, err := gateway.New(gateway.Config{
		ListenAddr: ":0", // 随机端口
		Targets: map[string]string{
			"/api": apiServer.URL,
		},
	})
	assert.NoError(t, err)

	// 启动网关
	gwServer := httptest.NewServer(gw.engine)
	defer gwServer.Close()

	// 创建客户端
	c := client.New(client.Config{
		GatewayAddr: gwServer.URL,
		Timeout:     5 * time.Second,
	})

	// 测试场景
	tests := []struct {
		name      string
		path      string
		algorithm limiter.Algorithm
		window    time.Duration
		limit     int64
		requests  int
		expected  struct {
			success int
			limited int
		}
	}{
		{
			name:      "Token Bucket Normal",
			path:      "/api/test1",
			algorithm: limiter.TokenBucket,
			window:    time.Minute,
			limit:     10,
			requests:  8,
			expected: struct {
				success int
				limited int
			}{
				success: 8,
				limited: 0,
			},
		},
		{
			name:      "Sliding Window Limit",
			path:      "/api/test2",
			algorithm: limiter.SlidingWindow,
			window:    time.Second,
			limit:     5,
			requests:  10,
			expected: struct {
				success int
				limited int
			}{
				success: 5,
				limited: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置限流规则
			err := c.SetRule(client.RuleConfig{
				Path:       tt.path,
				Algorithm:  tt.algorithm,
				WindowSize: tt.window,
				Limit:      tt.limit,
			})
			assert.NoError(t, err)

			// 发送请求并统计结果
			success := 0
			limited := 0

			for i := 0; i < tt.requests; i++ {
				resp, err := c.Get(tt.path)
				assert.NoError(t, err)

				if resp.StatusCode == http.StatusOK {
					success++
				} else if resp.StatusCode == http.StatusTooManyRequests {
					limited++
				}
				resp.Body.Close()
			}

			// 验证结果
			assert.Equal(t, tt.expected.success, success, "成功请求数不匹配")
			assert.Equal(t, tt.expected.limited, limited, "被限流请求数不匹配")
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	// 创建测试API服务器
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // 模拟处理时间
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	// 创建网关
	gw, err := gateway.New(gateway.Config{
		ListenAddr: ":0",
		Targets: map[string]string{
			"/api": apiServer.URL,
		},
	})
	assert.NoError(t, err)

	// 启动网关
	gwServer := httptest.NewServer(gw.engine)
	defer gwServer.Close()

	// 创建客户端
	c := client.New(client.Config{
		GatewayAddr: gwServer.URL,
		Timeout:     5 * time.Second,
	})

	// 设置限流规则
	err = c.SetRule(client.RuleConfig{
		Path:       "/api/test",
		Algorithm:  limiter.TokenBucket,
		WindowSize: time.Second,
		Limit:      50,
	})
	assert.NoError(t, err)

	// 并发测试
	concurrency := 10
	requests := 100
	results := make(chan int, requests)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 发送并发请求
	for i := 0; i < requests; i++ {
		go func() {
			resp, err := c.Get("/api/test")
			if err != nil {
				results <- -1
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode
		}()
	}

	// 统计结果
	success := 0
	limited := 0
	failed := 0

	for i := 0; i < requests; i++ {
		select {
		case <-ctx.Done():
			t.Fatal("测试超时")
		case status := <-results:
			switch status {
			case http.StatusOK:
				success++
			case http.StatusTooManyRequests:
				limited++
			default:
				failed++
			}
		}
	}

	// 验证结果
	t.Logf("成功: %d, 限流: %d, 失败: %d", success, limited, failed)
	assert.True(t, success > 0, "应该有成功的请求")
	assert.True(t, limited > 0, "应该有被限流的请求")
	assert.Equal(t, 0, failed, "不应该有失败的请求")
}
