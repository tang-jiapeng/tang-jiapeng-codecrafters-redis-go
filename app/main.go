package main

import (
	"flag"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	// 定义 --port 参数，默认 6379
	port := flag.Int("port", 6379, "port to listen on")
	flag.Parse()

	address := fmt.Sprintf("0.0.0.0:%d", *port)

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go commands.HandleConnection(conn)
	}

}
