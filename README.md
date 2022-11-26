# http-little-toy

一个简单的 `http` 并发测试工具。

#### README.md

[⭐️README.MD⭐️ En](README.en.md)

#### 仓库地址

如果喜欢就star⭐️一下吧，让它沉睡在你的收藏库里。

 [https://github.com/leihenshang/http-little-toy](https://github.com/leihenshang/http-little-toy)

 [https://gitee.com/leihenshang/http-little-toy](https://gitee.com/leihenshang/http-little-toy)

#### 介绍

灵感来源于 `github` 上各种版本的 `wrk` http并发测试工具，有一天看了一个go写的版本，就这？我也能行啊。我自己也造一个轮子吧。orz.

造轮子好玩吗？
真好玩！

#### todo

- [x] 命令行中加入设置 `header` 头

- [x] 命令行中加入设置 `body` 负载

- [ ] 完善一下 `request.json` 请求文件的逻辑

- [x] 添加记录响应数据日志的功能，方便分析

#### 使用

一般使用 -d 控制请求时间(秒),-t 控制线程数（当做用户数量来理解）就可以了。

还能使用request.json文件，你不用重新编写命令参数了,不过可能还有点小问题，在完善中。

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

#### 安装教程

1. 直接使用 `go install github.com/leihenshang/http-little-toy` ,再把你的`go/bin`放到环境变量里，使用 `http-little-toy` 带上参数，起飞吧，骚年。

2. 手动编译成二进制文件直接运行，可以放到全局变量中直接从命令行中执行。

#### 手动编译

```bash
# 把项目编译成可执行文件并输出到当前目录
go build -o httpToy
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
# 使用test-server
 go run . -u http://localhost:9090 -H aaa:bbbb -H ccc:ddd -body "hhhhh2333333" -d 2 -t 1
```

```bash
# 测试使用SSL,添加证书验证
go run . -u https://localhost:9090 -clientCert="./client-cert/ca.crt" -clientKey="./client-cert/ca.key" -caCert="./client-cert/ca.crt" -skipVerify=true
```

```txt
（1）生成客户端私钥 （生成CA私钥）
openssl genrsa -out ca.key 2048  //2048为长度

（2）生成CA证书
openssl req -x509 -new -nodes -key ca.key -subj "/CN=tonybai.com" -days 5000 -out ca.crt

接下来，生成server端的私钥，生成数字证书请求，并用我们的ca私钥签发server的数字证书：

（1）生成服务端私钥
openssl genrsa -out server.key 2048 //2048为长度

（2）生成证书请求文件
openssl req -new -key server.key -subj "/CN=localhost" -out server.csr

（3）根据CA的私钥和上面的证书请求文件生成服务端证书
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 5000

```