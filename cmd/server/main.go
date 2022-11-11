package main

import (
	"flag"
	proxy "go-socks5-proxy"
)

func main() {
	address := flag.String("address", ":7777", "代理服务监听地址")
	flag.Parse()
	proxy.Server(*address)
}
