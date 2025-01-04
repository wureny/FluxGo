package blackbox

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	// 创建测试API服务器
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "ok", "path": "%s"}`, r.URL.Path)
	}))
	defer apiServer.Close()

	// ... 其余测试代码 ...
}
