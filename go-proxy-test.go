package main

import (
	"fmt"
	"net"
)

func main() {
	var b [1024]byte
	resp, err := net.Dial("tcp", "localhost:1023")
	if err != nil {
		panic(err)
	}
	resp.Write([]byte("GET http://www.wjlin0.com/ HTTP/1.1\nHost: www.wjlin0.com\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:71.0) Gecko/20100101 Firefox/71.0\nAccept-Encoding: gzip"))
	resp.Read(b[:])
	fmt.Println(string(b[:]))
}
