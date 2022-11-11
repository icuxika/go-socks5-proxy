package main

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
)

// ProtocolVersion 版本和认证交互阶段客户端向服务端发送的报文
type ProtocolVersion struct {
	VER      uint8   // SOCKS版本
	NMETHODS uint8   // 支持的认证方式个数
	METHODS  []uint8 // 支持的认证方式集合
}

// Socks5Resolution 数据交互阶段客户端向服务端发送的报文
type Socks5Resolution struct {
	VER        uint8        // SOCKS版本
	CMD        uint8        // SOCK的命令码，[0x01 表示CONNECT请求]，[0x02 表示BIND请求]，[0x03 表示UDP转发]
	RSV        uint8        // 0x00，保留
	ATYP       uint8        // DST_ADDR类型，[0x01 IPv4地址]，[0x03 域名类型]，[0x04 IPv6地址]
	DST_ADDR   []byte       // 目标服务地址，如果是IPv4类型，则固定4个字节长度；如果是域名类型，第一个字节是域名长度，剩余的内容为域名内容；如果是IPv6类型，固定16个字节长度
	DST_PORT   uint16       // 目标服务端口，固定两个字节长度
	DST_DOMAIN string       // ATYP为0x03时记录域名
	RAW_ADDR   *net.TCPAddr // 最终访问目标服务器使用的IP:PORT
}

// 版本和认证交互阶段
func (p *ProtocolVersion) handleHandshake(b []byte) ([]byte, error) {
	n := len(b)
	if n < 3 {
		return nil, errors.New("协议错误，NMETHODS不对")
	}
	p.VER = b[0]
	if p.VER != 0x05 {
		return nil, errors.New("协议错误，仅支持Socks5")
	}
	p.NMETHODS = b[1]
	if n != int(2+p.NMETHODS) {
		return nil, errors.New("协议错误，NMETHODS不对")
	}
	p.METHODS = b[2 : 2+p.NMETHODS]
	useMethod := byte(0x00)
	for _, v := range p.METHODS {
		if v == 0x00 {
			useMethod = 0x00
		}
	}
	if useMethod != 0x00 {
		return nil, errors.New("协议错误，加密方法不对")
	}

	log.Printf("入1：版本：%v\n", p.VER)
	log.Printf("入1：认证方式个数：%v\n", p.NMETHODS)
	log.Printf("入1：认证方式：%v\n", p.METHODS)

	response := []byte{0x05, useMethod}
	return response, nil
}

// 数据交互阶段
func (s *Socks5Resolution) handleRequest(b []byte) ([]byte, error) {
	n := len(b)
	if n < 7 {
		return nil, errors.New("请求协议错误")
	}
	s.VER = b[0]
	if s.VER != 0x05 {
		return nil, errors.New("该协议不是Socks5协议")
	}
	s.CMD = b[1]
	if s.CMD != 1 {
		return nil, errors.New("客户端请求类型不为代理连接，其他功能暂时不支持")
	}
	s.RSV = b[2]
	s.ATYP = b[3]

	switch s.ATYP {
	case 1:
		s.DST_ADDR = b[4 : 4+net.IPv4len]
	case 3:
		s.DST_DOMAIN = string(b[5 : n-2])
		ipAddr, err := net.ResolveIPAddr("ip", s.DST_DOMAIN)
		if err != nil {
			return nil, err
		}
		s.DST_ADDR = ipAddr.IP
	case 4:
		s.DST_ADDR = b[4 : 4+net.IPv6len]
	default:
		return nil, errors.New("IP地址错误")
	}

	s.DST_PORT = binary.BigEndian.Uint16(b[n-2 : n])

	// DST_ADDR转换成IP地址，防止DNS污染和封杀
	s.RAW_ADDR = &net.TCPAddr{
		IP:   s.DST_ADDR,
		Port: int(s.DST_PORT),
	}

	log.Printf("入2：目标服务器地址类型：%v\n", s.ATYP)
	log.Printf("入2：目标服务器域名：%v\n", s.DST_DOMAIN)
	log.Printf("入2：目标服务器地址：%v\n", s.DST_ADDR)
	log.Printf("入2：目标服务器端口：%v\n", s.DST_PORT)
	log.Printf("入2：最终链接目标服务器使用的地址：%v\n", s.RAW_ADDR)

	response := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	return response, nil
}
