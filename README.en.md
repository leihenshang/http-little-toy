# http-little-toy

A simple http concurrent testing tool.

#### repository

 [https://github.com/leihenshang/http-little-toy](https://github.com/leihenshang/http-little-toy)

 [https://gitee.com/leihenshang/http-little-toy](https://gitee.com/leihenshang/http-little-toy)

#### introduce

Inspiration comes from 'github' on various versions of 'wrk' http concurrent testing tools。

Is it fun to build wheels?
That was fun!
orz.

#### feature

- [x] Command line directives support setting `http header`

- [x] Command line directives support setting `http body`

- [x] Prefect the `request.json` sample file

- [x] Adding a response log

#### instructions

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
        -f 
                 specify the request definition file. --default="".
        -gen 
                 generate the request definition file template to the current directory. --default=false.
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

#### installation tutorial

1. Use `go install github.com/leihenshang/http-little-toy` install to `go bin`.

2. Manual compile.

#### manual compile

```bash
# 把项目编译成可执行文件并输出到当前目录
go build -o httpToy
```

#### execute

```bash
# 使用纯命令
 ./httpToy -d 10 -t 80 -u http://127.0.0.1:9090

# or

# 使用请求文件
./httpToy -d 10 -t 80 -f request_sample.json

```

```bash
# 使用test-server
 go run . -u http://localhost:9090 -H aaa:bbbb -H ccc:ddd -body "hhhhh2333333" -d 2 -t 1
```
