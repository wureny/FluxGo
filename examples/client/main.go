package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/wureny/FluxGo/internal/limiter"
	"github.com/wureny/FluxGo/pkg/client"
)

var (
	gatewayAddr = flag.String("gateway", "http://localhost:8080", "网关地址")
	concurrent  = flag.Int("concurrent", 5, "并发数")
	total       = flag.Int("total", 100, "总请求数")
)

func main() {
	flag.Parse()

	// 创建客户端
	c := client.New(client.Config{
		GatewayAddr: *gatewayAddr,
		Timeout:     5 * time.Second,
	})

	// 设置限流规则示例
	rules := []struct {
		path      string
		algorithm limiter.Algorithm
		window    time.Duration
		limit     int64
	}{
		{
			path:      "/api/v1/users",
			algorithm: limiter.TokenBucket,
			window:    time.Minute,
			limit:     100,
		},
		{
			path:      "/api/v1/orders",
			algorithm: limiter.SlidingWindow,
			window:    time.Second,
			limit:     10,
		},
		{
			path:      "/api/v2/products",
			algorithm: limiter.LeakyBucket,
			window:    time.Minute,
			limit:     50,
		},
	}

	// 设置限流规则
	for _, rule := range rules {
		if err := c.SetRule(client.RuleConfig{
			Path:       rule.path,
			Algorithm:  rule.algorithm,
			WindowSize: rule.window,
			Limit:      rule.limit,
		}); err != nil {
			log.Fatalf("设置限流规则失败: path=%s, error=%v", rule.path, err)
		}
		log.Printf("设置限流规则: path=%s, algorithm=%s, window=%s, limit=%d",
			rule.path, rule.algorithm, rule.window, rule.limit)
	}

	// 测试限流效果
	testPaths := []string{
		"/api/v1/users",
		"/api/v1/orders",
		"/api/v2/products",
	}

	for _, path := range testPaths {
		log.Printf("\n开始测试路径: %s", path)
		testRateLimit(c, path)
	}
}

// 测试限流效果
func testRateLimit(c *client.Client, path string) {
	var (
		wg        sync.WaitGroup
		success   int32
		failed    int32
		startTime = time.Now()
	)

	// 创建工作通道
	jobs := make(chan int, *total)
	for i := 0; i < *total; i++ {
		jobs <- i
	}
	close(jobs)

	// 启动工作协程
	for i := 0; i < *concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				resp, err := c.Get(path)
				if err != nil {
					log.Printf("请求失败: %v", err)
					continue
				}

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				if resp.StatusCode == 200 {
					success++
					log.Printf("请求成功: status=%d, body=%s", resp.StatusCode, string(body))
				} else {
					failed++
					log.Printf("请求失败: status=%d, body=%s", resp.StatusCode, string(body))
				}

				// 稍微延迟一下，避免请求太快
				time.Sleep(time.Millisecond * 10)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 打印统计信息
	fmt.Printf("\n测试结果 - %s:\n", path)
	fmt.Printf("总请求数: %d\n", *total)
	fmt.Printf("成功请求: %d\n", success)
	fmt.Printf("被限流数: %d\n", failed)
	fmt.Printf("总耗时: %s\n", duration)
	fmt.Printf("平均QPS: %.2f\n", float64(*total)/duration.Seconds())
}
