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
)

var (
	port int
	max  int
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
	flag.IntVar(&port, "port", 1024, "端口")
	flag.IntVar(&max, "max", 1024, "报文大小")
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
		go handleConnection(client)
	}
}

func checkArgs() {
	//flag.Usage()
	if max < 1024 {
		log.Panic("报文大小最好不小于1024个字节")
	}
}

func handleConnection(client net.Conn) {
	var b = make([]byte, max)
	_, err := client.Read(b[:])

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
	server, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return
	}
	if method == "CONNECT" {
		_, err := client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			return
		}
	} else {
		server.Write(b[:])
	}
	go io.Copy(server, client)
	go io.Copy(client, server)
	fmt.Printf("%v -> %v -> %v\n", client.LocalAddr(), server.LocalAddr(), serverAddress)
}
