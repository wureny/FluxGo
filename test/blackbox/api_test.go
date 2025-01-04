package blackbox

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/wureny/FluxGo/internal/gateway"
	"github.com/wureny/FluxGo/internal/limiter"
	"github.com/wureny/FluxGo/pkg/client"
)

func init() {
	// 设置gin为发布模式，减少日志输出
	gin.SetMode(gin.ReleaseMode)
}

// 测试基本的限流功能
func TestBasicRateLimit(t *testing.T) {
	// 创建测试API服务器
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer testServer.Close()

	// 创建网关
	gw, err := gateway.New(gateway.Config{
		ListenAddr: ":0",
		Targets: map[string]string{
			"/api": testServer.URL,
		},
	})
	assert.NoError(t, err)

	// 启动网关
	gwServer := httptest.NewServer(gw.GetHandler())
	defer gwServer.Close()

	// 创建客户端
	c := client.New(client.Config{
		GatewayAddr: gwServer.URL, // 使用测试服务器的URL
		Timeout:     5 * time.Second,
	})

	// 设置限流规则
	err = c.SetRule(client.RuleConfig{
		Path:       "/api/test",
		Algorithm:  limiter.TokenBucket,
		WindowSize: time.Second,
		Limit:      5,
	})
	assert.NoError(t, err)

	// 发送请求并验证限流效果
	for i := 0; i < 10; i++ {
		resp, err := c.Get("/api/test")
		if !assert.NoError(t, err) {
			t.Logf("请求失败: %v", err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		t.Logf("请求 %d 响应: status=%d, body=%s", i, resp.StatusCode, string(body))
		resp.Body.Close()

		if i < 5 {
			assert.Equal(t, http.StatusOK, resp.StatusCode, "前5个请求应该成功")
		} else {
			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "后续请求应该被限流")
		}
	}
}

// 测试并发请求下的限流效果
func TestConcurrentRateLimit(t *testing.T) {
	// 创建测试服务器
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // 模拟处理延迟
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer testServer.Close()

	// 创建和启动网关
	gw, err := gateway.New(gateway.Config{
		ListenAddr: ":0",
		Targets: map[string]string{
			"/api": testServer.URL,
		},
	})
	assert.NoError(t, err)

	gwServer := httptest.NewServer(gw.GetHandler())
	defer gwServer.Close()

	// 创建客户端并设置限流规则
	c := client.New(client.Config{
		GatewayAddr: gwServer.URL,
		Timeout:     5 * time.Second,
	})

	err = c.SetRule(client.RuleConfig{
		Path:       "/api/test",
		Algorithm:  limiter.TokenBucket,
		WindowSize: time.Second,
		Limit:      10,
	})
	assert.NoError(t, err)

	// 添加调试日志
	t.Logf("网关地址: %s", gwServer.URL)
	t.Logf("目标服务器地址: %s", testServer.URL)

	// 并发测试
	var (
		wg           sync.WaitGroup
		successCount int32
		limitCount   int32
		totalReqs    = 30
		concurrent   = 10
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < totalReqs/concurrent; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					resp, err := c.Get("/api/test")
					if err != nil {
						t.Logf("请求错误: %v", err)
						continue
					}
					body, _ := io.ReadAll(resp.Body)
					t.Logf("响应: status=%d, body=%s", resp.StatusCode, string(body))
					resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
						atomic.AddInt32(&successCount, 1)
					} else if resp.StatusCode == http.StatusTooManyRequests {
						atomic.AddInt32(&limitCount, 1)
					}
				}
			}
		}()
	}

	wg.Wait()

	// 验证结果
	t.Logf("成功请求: %d, 被限流请求: %d", successCount, limitCount)
	assert.Equal(t, int32(totalReqs), successCount+limitCount, "总请求数应该匹配")
	assert.True(t, limitCount > 0, "应该有请求被限流")
	assert.True(t, successCount > 0, "应该有请求成功")
}
