package gateway

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/wureny/FluxGo/internal/limiter"
)

/*
- 限流中间件：
对所有非管理API的请求进行限流检查
使用客户端IP作为限流key
当请求被限流时返回429状态码
- 管理API：
POST /admin/rules：添加限流规则
DELETE /admin/rules/path：删除限流规则
GET /admin/rules/path：获取限流规则
- 反向代理：
将请求转发到配置的目标服务器
支持基于路径前缀的路由
- 配置灵活：
支持配置监听地址
支持配置多个目标服务器
*/

// Gateway API网关结构体
type Gateway struct {
	// 限流规则管理器
	ruleManager *limiter.RuleManager
	// 路由引擎
	engine *gin.Engine
	// 目标服务器地址映射
	targets map[string]*url.URL
}

// Config 网关配置
type Config struct {
	// 监听地址
	ListenAddr string
	// 目标服务器地址映射 (路径前缀 -> 目标URL)
	Targets map[string]string
}

// New 创建新的API网关
func New(config Config) (*Gateway, error) {
	g := &Gateway{
		ruleManager: limiter.NewRuleManager(),
		engine:      gin.Default(),
		targets:     make(map[string]*url.URL),
	}

	// 解析并存储目标服务器URL
	for path, target := range config.Targets {
		targetURL, err := url.Parse(target)
		if err != nil {
			return nil, fmt.Errorf("invalid target URL for path %s: %v", path, err)
		}
		g.targets[path] = targetURL
	}

	// 设置中间件和路由
	g.setupRoutes()

	return g, nil
}

// setupRoutes 设置路由和中间件
func (g *Gateway) setupRoutes() {
	// 限流中间件
	g.engine.Use(g.rateLimitMiddleware())

	// 管理API
	admin := g.engine.Group("/admin")
	{
		admin.POST("/rules", g.addRule)
		admin.DELETE("/rules/*path", g.removeRule)
		admin.GET("/rules/*path", g.getRule)
	}

	// 所有其他请求都转发到目标服务器
	g.engine.NoRoute(g.handleProxy)
}

// rateLimitMiddleware 限流中间件
func (g *Gateway) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过管理API的限流
		if len(c.Request.URL.Path) >= 6 && c.Request.URL.Path[:6] == "/admin" {
			c.Next()
			return
		}

		// 使用客户端IP作为限流key
		key := c.ClientIP()

		// 检查是否允许请求通过
		allowed, waitTime := g.ruleManager.Allow(c, c.Request.URL.Path, key)
		if !allowed {
			c.Header("X-RateLimit-Retry-After", fmt.Sprintf("%d", int64(waitTime.Seconds())))
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}

// handleProxy 处理代理请求
func (g *Gateway) handleProxy(c *gin.Context) {
	path := c.Request.URL.Path
	var targetURL *url.URL

	// 添加调试日志
	log.Printf("收到请求: path=%s", path)

	// 查找匹配的目标服务器
	for prefix, target := range g.targets {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			targetURL = target
			log.Printf("找到目标服务器: prefix=%s, target=%s", prefix, target)
			break
		}
	}

	if targetURL == nil {
		log.Printf("未找到匹配的目标服务器: path=%s", path)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(c.Writer, c.Request)
}

// addRule 添加限流规则
func (g *Gateway) addRule(c *gin.Context) {
	var rule limiter.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	if err := g.ruleManager.AddRule(path, rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// removeRule 移除限流规则
func (g *Gateway) removeRule(c *gin.Context) {
	path := c.Param("path")
	g.ruleManager.RemoveRule(path)
	c.Status(http.StatusOK)
}

// getRule 获取限流规则
func (g *Gateway) getRule(c *gin.Context) {
	path := c.Param("path")
	rule, exists := g.ruleManager.GetRule(path)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// Run 启动API网关
func (g *Gateway) Run(addr string) error {
	return g.engine.Run(addr)
}

// Close 关闭API网关
func (g *Gateway) Close() error {
	return g.ruleManager.Close()
}

// GetHandler 返回HTTP处理器
func (g *Gateway) GetHandler() http.Handler {
	return g.engine
}
