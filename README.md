# http-little-toy

#### 介绍
一个简单的http基准测试工具。
灵感来自于 github 上的各种叫做wrk的工具，自己也造一个轮子玩一玩。

#### 软件架构
使用golang开发，使用协程的特性实现并发请求测试。


#### 安装教程
1. 编译成二进制文件直接运行，还可以放到全局变量中从而直接从命令行中执行。

#### 使用说明

```bash

// 使用示例
go run . -d 10 -t 80 -f request_sample.json

```

#### 参与贡献

1. 开发中，暂无。


#### 特技

1.  使用 Readme\_XXX.md 来支持不同的语言，例如 Readme\_en.md, Readme\_zh.md
2.  Gitee 官方博客 [blog.gitee.com](https://blog.gitee.com)
3.  你可以 [https://gitee.com/explore](https://gitee.com/explore) 这个地址来了解 Gitee 上的优秀开源项目
4.  [GVP](https://gitee.com/gvp) 全称是 Gitee 最有价值开源项目，是综合评定出的优秀开源项目
5.  Gitee 官方提供的使用手册 [https://gitee.com/help](https://gitee.com/help)
6.  Gitee 封面人物是一档用来展示 Gitee 会员风采的栏目 [https://gitee.com/gitee-stars/](https://gitee.com/gitee-stars/)
