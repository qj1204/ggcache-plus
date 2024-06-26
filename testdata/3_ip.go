package main

import (
	"fmt"
	"net"
)

func main() {
	ipStr := "127.0.0.1"
	ip := net.ParseIP(ipStr)
	if ipStr == "localhost" || ip != nil {
		fmt.Println(true)
	} else {
		fmt.Println(false)
	}
}
