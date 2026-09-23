# 跨平台编译指南 Cross-platform Build Guide

> 返回 [README](../README.md) · Back to [README](../README.md)

编译产物为零 CGO 依赖的单二进制，可直接拷贝到目标平台运行。
Binaries are CGO-free single files and can be copied directly to the target platform.

## Linux & macOS Shell 编译 Build

```bash
# Linux (amd64)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/http-little-toy

# Linux (arm64)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/http-little-toy

# macOS (amd64)
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/http-little-toy

# macOS (Apple Silicon)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/http-little-toy

# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/http-little-toy.exe
```

## Windows CMD 编译 Build

```cmd
REM Linux编译 Linux build
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o bin/http-little-toy

REM macOS编译 macOS build
SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build -o bin/http-little-toy

REM Windows编译 Windows build
go build -o bin/http-little-toy.exe
```

## 开发调试 Development

```bash
# 直接运行 Run directly
go run . -u http://localhost:9090 -t 10 -d 10

# 配合 test-server 本地联调 Test against the bundled local server
go run ./test-server
```

详见 [test-server README](../test-server/README.md)。
See [test-server README](../test-server/README.md) for details.
