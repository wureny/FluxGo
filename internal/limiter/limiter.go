package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wureny/FluxGo/internal/algorithms"
	"github.com/wureny/FluxGo/internal/algorithms/leakybucket"
	"github.com/wureny/FluxGo/internal/algorithms/slidinglog"
	"github.com/wureny/FluxGo/internal/algorithms/slidingwindow"
	"github.com/wureny/FluxGo/internal/algorithms/tokenbucket"
)

// Algorithm 限流算法类型
type Algorithm string

const (
	SlidingLog    Algorithm = "sliding_log"
	SlidingWindow Algorithm = "sliding_window"
	LeakyBucket   Algorithm = "leaky_bucket"
	TokenBucket   Algorithm = "token_bucket"
)

// Rule 限流规则
type Rule struct {
	// 限流算法类型
	Algorithm Algorithm
	// 限流配置
	Config algorithms.Config
}

// RuleManager 限流规则管理器
type RuleManager struct {
	mu sync.RWMutex
	// 路径 -> 限流规则的映射
	rules map[string]Rule
	// 路径 -> 限流器实例的映射
	limiters map[string]algorithms.RateLimiter
}

// NewRuleManager 创建新的规则管理器
func NewRuleManager() *RuleManager {
	return &RuleManager{
		rules:    make(map[string]Rule),
		limiters: make(map[string]algorithms.RateLimiter),
	}
}

// AddRule 添加限流规则
func (rm *RuleManager) AddRule(path string, rule Rule) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 创建对应的限流器实例
	limiter, err := rm.createLimiter(rule)
	if err != nil {
		return err
	}

	// 如果已存在旧的限流器，先关闭它
	if oldLimiter, exists := rm.limiters[path]; exists {
		oldLimiter.Close()
	}

	rm.rules[path] = rule
	rm.limiters[path] = limiter
	return nil
}

// RemoveRule 移除限流规则
func (rm *RuleManager) RemoveRule(path string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if limiter, exists := rm.limiters[path]; exists {
		limiter.Close()
		delete(rm.limiters, path)
	}
	delete(rm.rules, path)
}

// Allow 判断请求是否允许通过
func (rm *RuleManager) Allow(ctx context.Context, path string, key string) (bool, time.Duration) {
	rm.mu.RLock()
	limiter, exists := rm.limiters[path]
	rm.mu.RUnlock()

	if !exists {
		// 如果路径没有配置限流规则，默认允许通过
		return true, 0
	}

	return limiter.Allow(ctx, key)
}

// GetRule 获取指定路径的限流规则
func (rm *RuleManager) GetRule(path string) (Rule, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	rule, exists := rm.rules[path]
	return rule, exists
}

// createLimiter 根据规则创建对应的限流器实例
func (rm *RuleManager) createLimiter(rule Rule) (algorithms.RateLimiter, error) {
	switch rule.Algorithm {
	case SlidingLog:
		return slidinglog.NewLimiter(rule.Config), nil
	case SlidingWindow:
		return slidingwindow.NewLimiter(rule.Config), nil
	case LeakyBucket:
		return leakybucket.NewLimiter(rule.Config), nil
	case TokenBucket:
		return tokenbucket.NewLimiter(rule.Config), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", rule.Algorithm)
	}
}

// Close 关闭所有限流器实例
func (rm *RuleManager) Close() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, limiter := range rm.limiters {
		if err := limiter.Close(); err != nil {
			return err
		}
	}

	rm.limiters = make(map[string]algorithms.RateLimiter)
	rm.rules = make(map[string]Rule)
	return nil
}
