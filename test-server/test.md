# test

#### SSL 测试

```bash

# 服务端ssl
go run . -certFile="../client-cert/server.crt" -keyFile="../client-cert/server.key"

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