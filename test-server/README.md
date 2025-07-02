# test

#### SSL 测试

```bash

# 服务端TLS启动
go run . -certFile="../client-cert/server.crt" -keyFile="../client-cert/server.key"

# 服务端TLS启动,并启用HTTP2
go run main.go  -certFile="server.crt" -keyFile="server.key" -h2

# little-toy测试使用SSL,添加证书验证
go run . -u https://localhost:9090 -clientCert="./client-cert/ca.crt" -clientKey="./client-cert/ca.key" -caCert="./client-cert/ca.crt" -skipVerify=true
```

```txt
（1）生成客户端私钥 （生成CA私钥）
openssl genrsa -out ca.key 2048  //2048为长度

（2）生成CA证书

openssl req -x509 -new -nodes -key ca.key -subj "/CN=toy.com" -days 5000 -out ca.crt

接下来，生成server端的私钥，生成数字证书请求，并用我们的ca私钥签发server的数字证书：

（1）生成服务端私钥
openssl genrsa -out server.key 2048 //2048为长度

（2）生成证书请求文件
# 在 git bash 下(windows),/CN 写成  //CN 转义一下,否则会报错
openssl req -new -key server.key -subj "/CN=localhost" -out server.csr

（3）根据CA的私钥和上面的证书请求文件生成服务端证书
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 5000

```

## 生成自签名证书
   ```
    #生成私钥  
   openssl genrsa -out server.key 2048

   # 生成自签名证书（有效期 365 天）
openssl req -new -x509 -key server.key -out server.crt -days 365 -subj "/CN=localhost"
   ```



## 使用 curl 测试 

```
# -v 打印其他信息
# -k 跳过证书验证,正式环境注意
curl --http2 -k https://localhost:9090

```