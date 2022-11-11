package main

import (
	"io"
	"log"
	"net"
	"sync"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":7777")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("服务器监听端口：%v\n", tcpAddr)
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn *net.TCPConn) {
	if conn == nil {
		return
	}

	log.Println("-------------------- 版本和认证方式交互 --------------------")

	buff := make([]byte, 255)
	n, err := conn.Read(buff)
	if err != nil {
		log.Fatal(err)
		return
	}

	var protocolVersion ProtocolVersion
	response, err := protocolVersion.handleHandshake(buff[0:n])
	_, err = conn.Write(response)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("-------------------- 数据交互 --------------------")

	buff = make([]byte, 255)
	n, err = conn.Read(buff)
	if err != nil {
		log.Fatal(err)
		return
	}
	var socks5Resolution Socks5Resolution
	// 此方法目前在与目标服务器建立连接之前已经回应客户端连接成功
	response, err = socks5Resolution.handleRequest(buff[0:n])
	_, err = conn.Write(response)
	if err != nil {
		log.Fatal(err)
		return
	}

	// 链接真正的远程服务
	dstServer, err := net.DialTCP("tcp", nil, socks5Resolution.RAW_ADDR)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dstServer.Close()

	log.Println("-------------------- 握手结束 --------------------")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// 本地客户端数据拷贝到目标服务器
	go func() {
		defer wg.Done()
		copy(conn, dstServer)
	}()

	// 目标服务器数据拷贝到本地客户端
	go func() {
		defer wg.Done()
		copy(dstServer, conn)
	}()

	wg.Wait()
}

func copy(src io.ReadWriteCloser, dst io.ReadWriteCloser) (written int64, err error) {
	size := 1024
	buf := make([]byte, size)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
