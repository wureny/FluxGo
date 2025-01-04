package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wureny/FluxGo/internal/algorithms"
	"github.com/wureny/FluxGo/internal/limiter"
)

// Client FluxGo客户端
type Client struct {
	// 网关地址
	gatewayAddr string
	// HTTP客户端
	httpClient *http.Client
}

// Config 客户端配置
type Config struct {
	// 网关地址
	GatewayAddr string
	// HTTP客户端超时时间
	Timeout time.Duration
}

// RuleConfig 限流规则配置
type RuleConfig struct {
	// 路径
	Path string
	// 限流算法
	Algorithm limiter.Algorithm
	// 时间窗口大小
	WindowSize time.Duration
	// 限制次数
	Limit int64
}

// New 创建新的客户端
func New(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &Client{
		gatewayAddr: config.GatewayAddr,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SetRule 设置限流规则
func (c *Client) SetRule(config RuleConfig) error {
	rule := limiter.Rule{
		Algorithm: config.Algorithm,
		Config: algorithms.Config{
			WindowSize: config.WindowSize,
			Limit:      config.Limit,
		},
	}

	body, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("marshal rule failed: %v", err)
	}

	// 构造完整的URL
	gatewayURL, err := url.Parse(c.gatewayAddr)
	if err != nil {
		return fmt.Errorf("invalid gateway address: %v", err)
	}

	// 构造请求URL
	reqURL := *gatewayURL
	reqURL.Path = "/admin/rules"
	reqURL.RawQuery = "path=" + url.QueryEscape(config.Path)

	req, err := http.NewRequest(http.MethodPost, reqURL.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set rule failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// GetRule 获取限流规则
func (c *Client) GetRule(path string) (*RuleConfig, error) {
	url := fmt.Sprintf("%s/admin/rules/%s", c.gatewayAddr, path)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get rule failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var rule limiter.Rule
	if err := json.NewDecoder(resp.Body).Decode(&rule); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	return &RuleConfig{
		Path:       path,
		Algorithm:  rule.Algorithm,
		WindowSize: rule.Config.WindowSize,
		Limit:      rule.Config.Limit,
	}, nil
}

// RemoveRule 删除限流规则
func (c *Client) RemoveRule(path string) error {
	url := fmt.Sprintf("%s/admin/rules/%s", c.gatewayAddr, path)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remove rule failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// Do 发送HTTP请求
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// 确保请求发送到网关
	gatewayURL, err := url.Parse(c.gatewayAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway address: %v", err)
	}

	req.URL.Scheme = gatewayURL.Scheme
	req.URL.Host = gatewayURL.Host
	return c.httpClient.Do(req)
}

// Get 发送GET请求
func (c *Client) Get(path string) (*http.Response, error) {
	// 构造完整的URL
	gatewayURL, err := url.Parse(c.gatewayAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway address: %v", err)
	}

	// 确保path以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 构造请求URL
	reqURL := *gatewayURL
	reqURL.Path = path

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

// Post 发送POST请求
func (c *Client) Post(path string, contentType string, body io.Reader) (*http.Response, error) {
	// 构造完整的URL
	gatewayURL, err := url.Parse(c.gatewayAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway address: %v", err)
	}

	// 确保path以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 构造请求URL
	reqURL := *gatewayURL
	reqURL.Path = path

	req, err := http.NewRequest(http.MethodPost, reqURL.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.httpClient.Do(req)
}
