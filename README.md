# http-little-toy

#### 介绍

一个 `http` 并发测试工具。

灵感来自于 github 上的各种叫做wrk的工具，自己也造一个轮子。

#### todo

- [x] 命令行中加入设置 `header` 头

- [x] 命令行中加入设置 `body` 负载

#### 使用

一般使用 -d 控制请求时间(秒),-t 控制线程数（当做用户数量来理解）就可以了。推荐使用request.json文件。

```bash
$ go run . -h
Usage: httpToy <options>
Options:
        -H 
                 The http header. default:[]
        -allowRedirects 
                 allowRedirects default:true
        -body 
                 The http body default:
        -caCert 
                 caCert default:
        -clientCert 
                 clientCert default:
        -clientKey 
                 clientKey default:
        -compression 
                 Use keep-alive for http protocol. default:true
        -d 
                 Duration of request.The unit is seconds. default:0
        -f 
                 specify the request definition file. default:
        -gen 
                 generate the request definition file template to the current directory. default:false
        -h 
                 show help tips default:false
        -keepAlive 
                 Use keep-alive for http protocol. default:true
        -skipVerify 
                 TLS skipVerify default:false
        -t 
                 Number of threads. default:0
        -timeOut 
                 the time out to wait response default:1000
        -u 
                 The URL you want to test default:
        -useHttp2 
                 useHttp2 default:false
        -v 
                 show app version. default:false

```

#### 安装教程

 编译成二进制文件直接运行，还可以放到全局变量中从而直接从命令行中执行。

编译:

```bash
# 把项目编译成可执行文件并输出到当前目录
go build -o httpToy
```

执行:

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
