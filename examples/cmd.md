1. 启动网关服务器 
go run cmd/server/main.go
2. 启动示例API服务器 
go run examples/server/main.go -port 8081
3. 运行测试客户端
go run examples/client/main.go -concurrent 5 -total 100