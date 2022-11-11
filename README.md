Go SOCKS5 Proxy
==========
简单的 SOCKS5 协议交互模拟

- [Socks5Proxy（主要copy目标）](https://github.com/shikanon/socks5proxy)
- [使用netty实现socks5协议](https://www.cnblogs.com/kuangdaoyizhimei/p/14735895.html)

## 使用

### Server

#### 运行

```shell
cd cmd/server
go run .
```

#### 帮助

```shell
go run . --help
```

#### 指定其他监听地址

```shell
go run . --address :6666
```
