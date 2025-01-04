package blackbox

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConcurrentRequests(t *testing.T) {
	// 创建测试API服务器
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // 模拟处理时间
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	// ... 其余测试代码 ...
}
