# http-little-toy

## 简介(INTRODUCTION)

 这是一个简单的 `http` 并发测试工具。如果喜欢它就点一下star⭐️吧，让它沉睡在你的收藏库。
 
 This is a sample HTTP concurrency testing tool. If you like it, click on star⭐️,  let it sleep in your collection.
 
## 使用(TUTORIAL)

```bash
$ http-little-toy -h
Usage: http-little-toy <options>Options:
        -allowRedirects 
                 allowRedirects. --default=true.
        -body 
                 The http body. --default="".
        -caCert 
                 caCert. --default="".
        -clientCert 
                 clientCert. --default="".
        -clientKey 
                 clientKey. --default="".
        -compression 
                 Use keep-alive for http protocol. --default=true.
        -d 
                 Duration of request.The unit is seconds. --default=10.
        -h 
                 show help tips. --default=false.
        -h2
                 useHttp2. --default=false.
        -header
                 The http header. --default=[].
        -keepAlive
                 Use keep-alive for http protocol. --default=true.
        -m
                 The http method. --default=GET.
        -skipVerify
                 TLS skipVerify. --default=false.
        -t
                 Number of threads. --default=10.
        -timeout
                 the time out to wait response.the unit is seconds. --default=5.
        -u
                 The URL you want to test. --default="".
        -v
                 show version. --default=false.


```

## 安装(INSTALLATION)

1. 有go运行时的话，执行 `go install github.com/leihenshang/http-little-toy`
    - If there is a GO runtime, execute `go install github.com/leihenshang/http-little-toy`
2. 确保 `go/bin` 目录在全局环境变量里,然后就可以使用 `http-little-toy` 执行测试了
    - Make sure the `go/bin` directory is in the global environment variable, then you can use `http-little-toy` to perform the test.
3. 或者，你也可以编译运行
    - Alternatively, you can also compile and run it.


## 编译 （COMPILE）

linux & mac 

```bash
# linux 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/http-little-toy

# windows 
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/http-little-toy.exe

# mac
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/http-little-toy
```

windows 

```cmd
# Mac
SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build -o bin/http-little-toy
 
# Linux
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o bin/http-little-toy

# window
go build -o bin/http-little-toy.exe
```

## 例子 （EXAMPLE）

```bash
# 直接使用 (direct use)
 ./http-little-toy -d 10 -t 80 -u http://127.0.0.1:9090

 # 带上header (with header)
 go run . -u http://localhost:9090 -H aaa:bbbb -H ccc:ddd -body "hhhhh2333333" -d 10 -t 10 

 # 复杂一点 （more complicated ）
 ./http-little-toy -u http://localhost:2025/aa/bb/cc/ee/ff/list -m POST -header token:itoken_1754640941655527771 -header Content-Type:application/json -body='{
  "pageSize": 100,
  "page": 1
}' -d 60
```