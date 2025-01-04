package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/wureny/FluxGo/internal/gateway"
	"github.com/wureny/FluxGo/internal/limiter"
	"github.com/wureny/FluxGo/pkg/client"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "配置文件路径")
}

// 配置结构体
type Config struct {
	Gateway struct {
		ListenAddr string            `mapstructure:"listen_addr"`
		Targets    map[string]string `mapstructure:"targets"`
	} `mapstructure:"gateway"`

	DefaultRules map[string]struct {
		Algorithm  string `mapstructure:"algorithm"`
		WindowSize string `mapstructure:"window_size"`
		Limit      int64  `mapstructure:"limit"`
	} `mapstructure:"default_rules"`
}

func main() {
	flag.Parse()

	// 加载配置
	var config Config
	if err := loadConfig(configFile, &config); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建网关
	gw, err := gateway.New(gateway.Config{
		ListenAddr: config.Gateway.ListenAddr,
		Targets:    config.Gateway.Targets,
	})
	if err != nil {
		log.Fatalf("创建网关失败: %v", err)
	}

	// 启动网关
	go func() {
		log.Printf("启动网关，监听地址: %s", config.Gateway.ListenAddr)
		if err := gw.Run(config.Gateway.ListenAddr); err != nil {
			log.Fatalf("网关运行失败: %v", err)
		}
	}()

	// 等待网关启动
	time.Sleep(time.Second * 2)

	// 创建客户端用于设置默认规则
	c := client.New(client.Config{
		GatewayAddr: "http://localhost" + config.Gateway.ListenAddr,
		Timeout:     5 * time.Second,
	})

	// 设置默认限流规则
	for path, rule := range config.DefaultRules {
		windowSize, err := time.ParseDuration(rule.WindowSize)
		if err != nil {
			log.Fatalf("解析窗口大小失败: path=%s, error=%v", path, err)
		}

		// 添加重试逻辑
		var setRuleErr error
		for i := 0; i < 3; i++ { // 最多重试3次
			setRuleErr = c.SetRule(client.RuleConfig{
				Path:       path,
				Algorithm:  limiter.Algorithm(rule.Algorithm),
				WindowSize: windowSize,
				Limit:      rule.Limit,
			})
			if setRuleErr == nil {
				break
			}
			time.Sleep(time.Second) // 重试前等待1秒
		}
		if setRuleErr != nil {
			log.Fatalf("设置默认规则失败: path=%s, error=%v", path, err)
		}

		log.Printf("设置默认规则: path=%s, algorithm=%s, window_size=%s, limit=%d",
			path, rule.Algorithm, rule.WindowSize, rule.Limit)
	}

	// 优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("收到关闭信号，开始优雅关闭...")
		cancel()
	}()

	// 等待关闭信号
	<-ctx.Done()

	// 关闭网关
	if err := gw.Close(); err != nil {
		log.Printf("关闭网关失败: %v", err)
	}
	log.Println("网关已关闭")
}

// 加载配置文件
func loadConfig(file string, config interface{}) error {
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}
	return nil
}
