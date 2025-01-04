# FluxGo

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue)
[![Go Report Card](https://goreportcard.com/badge/github.com/wureny/FluxGo)](https://goreportcard.com/report/github.com/wureny/FluxGo)

FluxGo is a high-performance, configurable API rate limiting middleware that supports multiple rate limiting algorithms and can be used as a standalone API gateway.

## ✨ Features

- 🚀 Multiple Rate Limiting Algorithms
  - Sliding Window Log
  - Sliding Window Counter
  - Leaky Bucket
  - Token Bucket
- 🔌 Flexible Configuration
  - Dynamic rate limit rules
  - Customizable parameters
  - Path-level rate limiting
- 🌐 API Gateway Features
  - Reverse proxy
  - Route forwarding
  - Middleware support
- 📊 Monitoring & Statistics
  - Request counting
  - Rate limit statistics
  - Wait time calculation

## 🚀 Quick Start

### Installation

```bash
go get github.com/wureny/FluxGo
```

### Usage

Basic Usage

1. Start the gateway server:
```bash
go run cmd/server/main.go -config configs/config.yaml
```
2. Configure rate limit rules:
```go
client := client.New(client.Config{
GatewayAddr: "http://localhost:8080",
})
err := client.SetRule(client.RuleConfig{
Path: "/api/test",
Algorithm: limiter.TokenBucket,
WindowSize: time.Minute,
Limit: 100,
})
```

See [examples/cmd.md](examples/cmd.md) for more details.


## 📖 Documentation

### Rate Limiting Algorithms

1. **Sliding Window Log**
   - Precise counting
   - Suitable for high-precision scenarios
   - Higher memory usage

2. **Sliding Window Counter**
   - Moderate precision
   - Low memory usage
   - Simple implementation

3. **Leaky Bucket**
   - Fixed outflow rate
   - Ideal for constant rate scenarios
   - No burst support

4. **Token Bucket**
   - Supports burst traffic
   - Average rate control
   - More complex implementation

### Configuration
```yaml
gateway:
listen_addr: ":8080"
targets:
"/api/v1": "http://localhost:8081"
default_rules:
"/api/v1/users":
algorithm: "token_bucket"
window_size: "1m"
limit: 100
```

## 🔧 Development

### Project Structure
```markdown
FluxGo/
├── cmd/ # Command-line entries
│ └── server/ # API gateway server
├── internal/ # Private code
│ ├── algorithms/ # Rate limiting algorithms
│ ├── gateway/ # API gateway implementation
│ └── limiter/ # Core rate limiting logic
├── pkg/ # Exportable packages
│ └── client/ # Client SDK
├── configs/ # Configuration files
├── examples/ # Usage examples
└── test/ # Test files
├── whitebox/ # White box tests
└── blackbox/ # Black box tests
```
### Running Tests
Run white box tests
```bash
go test -v ./test/whitebox/...
```
Run black box tests
```bash
go test -v ./test/blackbox/...
```