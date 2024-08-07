# http-little-toy

## 简介-INTRODUCTION

- 一个简单的 `http` 并发测试工具。
  - A simple HTTP concurrency testing tool.
- 如果喜欢它就star⭐️一下吧，让它沉睡在你的收藏库里。
  - If you like it,please star it and let it sleep in your repository!
- 造轮子真好玩！orz.
  - Building wheel is very funny! orz!
 
#### 使用-TUTORIAL

```bash
$ http-little-toy -h
Usage: httpToy <options>
Options:
        -H 
                 The http header. --default=[].
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
                 Duration of request.The unit is seconds. --default=0.
        -h 
                 show help tips. --default=false.
        -keepAlive 
                 Use keep-alive for http protocol. --default=true.
        -log 
                 record request log to file. default: './log' --default=false.
        -skipVerify 
                 TLS skipVerify. --default=false.
        -t 
                 Number of threads. --default=0.
        -timeOut 
                 the time out to wait response. --default=1000.
        -u 
                 The URL you want to test. --default="".
        -useHttp2 
                 useHttp2. --default=false.
        -v 
                 show app version. --default=false.

```

#### 安装教程

1. 直接使用 `go install github.com/leihenshang/http-little-toy` ,再把你的`go/bin`放到环境变量里，使用 `http-little-toy` 带上参数，起飞吧，骚年。

2. 手动编译成二进制文件直接运行，可以放到全局变量中直接从命令行中执行。

#### 手动编译

linux & mac 

```bash
# 把项目编译成可执行文件并输出到当前目录
go build -o http-little-toy


## 在linux或mac上编译

# linux 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/http-little-toy

# windows 
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/http-little-toy.exe

# mac
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/http-little-toy
```

在windows 上编译

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

#window
go build -o bin/http-little-toy.exe
```

#### 执行测试命令

```bash
# 使用纯命令
 ./httpToy -d 10 -t 80 -u http://127.0.0.1:9090

# or

# 使用请求文件
./httpToy -d 10 -t 80 -f request_sample.json

```

```bash
# 使用test-server 测试
 go run . -u http://localhost:9090 -H aaa:bbbb -H ccc:ddd -body "hhhhh2333333" -d 2 -t 1
```


```bash 
# Common directive
go run . -u http://localhost:9090 -H aaa:bbbb -H ccc:ddd -body "hhhhh2333333" -d 10 -t 10 -log=true

```