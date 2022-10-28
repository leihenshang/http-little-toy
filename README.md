# http-little-toy

#### 介绍
一个关于的http基准测试工具。
灵感来自于 github 上的各种叫做wrk的工具，自己也造一个轮子玩一玩。

#### 软件架构
由golang开发，使用Json作为配置文件。


#### 安装教程
1. 编译成二进制文件直接运行，还可以放到全局变量中从而直接从命令行中执行。

#### 使用说明

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
