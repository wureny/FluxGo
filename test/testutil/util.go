package testutil

import (
	"net/http"
	"net/http/httptest"
)

// CreateTestServer 创建测试API服务器
func CreateTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// CreateTestGateway 创建测试网关
func CreateTestGateway(config GatewayConfig) (*Gateway, *httptest.Server) {
	// ... 创建网关的通用代码 ...
}
