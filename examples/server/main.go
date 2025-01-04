package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	port = flag.Int("port", 8081, "服务器端口")
)

func main() {
	flag.Parse()

	r := gin.Default()

	// 用户API
	r.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "获取用户列表",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 订单API
	r.GET("/api/v1/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "获取订单列表",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 产品API
	r.GET("/api/v2/products", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "获取产品列表",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("启动API服务器，监听地址: %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器运行失败: %v", err)
	}
}
