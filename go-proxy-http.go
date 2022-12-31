package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"
)

var (
	port    int
	max     int
	timeout time.Duration
)

func Banner() {

	fmt.Println(`
           __.__  .__       _______   
__  _  __ |__|  | |__| ____ \   _  \  
\ \/ \/ / |  |  | |  |/    \/  /_\  \ 
 \     /  |  |  |_|  |   |  \  \_/   \
  \/\_/\__|  |____/__|___|  /\_____  /
      \______|            \/       \/ 
        go-proxy-http `)
}
func Init() {
	flag.IntVar(&port, "p", 1024, "端口")
	flag.IntVar(&max, "m", 1024, "报文大小")
	flag.DurationVar(&timeout, "t", 10*time.Second, "响应延迟")
}
func main() {
	Init()
	Banner()
	flag.Parse()
	checkArgs()
	log.Printf("代理端口: %v 最大报文大小: %v", port, max)
	con, err := net.Listen("tcp", fmt.Sprintf(":%v", port))

	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := con.Accept()
		if err != nil {
			continue
		}
		go handleConnection(proxyConn{c: client})
	}
}

func checkArgs() {
	//flag.Usage()
	if max < 1024 {
		log.Panic("报文大小最好不小于1024个字节")
	}
}

type proxyConn struct {
	c net.Conn
}

func (p proxyConn) Read(b []byte) (n int, err error) {
	err = p.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return 0, err
	}
	return p.c.Read(b)
}

func (p proxyConn) Write(b []byte) (n int, err error) {
	err = p.SetWriteDeadline(time.Now().Add(timeout))
	if err != nil {
		return 0, err
	}
	return p.c.Write(b)
}

func (p proxyConn) Close() error {
	return p.c.Close()
}

func (p proxyConn) LocalAddr() net.Addr {
	return p.c.LocalAddr()
}

func (p proxyConn) RemoteAddr() net.Addr {
	return p.c.RemoteAddr()
}

func (p proxyConn) SetDeadline(t time.Time) error {
	return p.c.SetDeadline(t)
}

func (p proxyConn) SetReadDeadline(t time.Time) error {
	return p.c.SetReadDeadline(t)
}

func (p proxyConn) SetWriteDeadline(t time.Time) error {
	return p.c.SetWriteDeadline(t)
}

type pNet interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

func handleConnection(client proxyConn) {
	var b = make([]byte, max)
	err := client.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return
	}
	err = client.SetDeadline(time.Now().Add(timeout))
	_, err = client.Read(b[:])

	if err != nil {
		return
	}
	var method, proxyUrl, httpProtocolVersion, serverAddress string
	//fmt.Println(string(b[:]))
	n := bytes.IndexByte(b[:], '\n')
	if n == -1 {
		fmt.Println("error protocol,need only http https")
		return
	}
	_, err = fmt.Sscanf(string(b[:n]), "%v %v %v", &method, &proxyUrl, &httpProtocolVersion)
	if err != nil {
		fmt.Println(err)
		return
	}
	urlParse, err := url.Parse(proxyUrl)
	if err != nil {
		return
	}

	if method == "CONNECT" {
		serverAddress = urlParse.Scheme + fmt.Sprintf(":%v", urlParse.Opaque)

	} else {
		if strings.Index(urlParse.Host, ":") == -1 { //host不带端口， 默认80
			serverAddress = urlParse.Host + ":80"
		} else {
			serverAddress = urlParse.Host
		}
	}
	server1, err := net.Dial("tcp", serverAddress)
	server := proxyConn{c: server1}
	if err != nil {
		return
	}
	if method == "CONNECT" {
		_, err := client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			return
		}
	} else {
		_, err := server.Write(b[:])
		if err != nil {
			return
		}
	}
	go io.Copy(server, client)
	go io.Copy(client, server)
	fmt.Printf("%v -> %v -> %v\n", client.LocalAddr(), server.LocalAddr(), serverAddress)
}
