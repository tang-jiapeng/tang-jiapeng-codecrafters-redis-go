package main

import (
	"flag"
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/commands"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	// 定义 --port 参数，默认 6379
	port := flag.Int("port", 6379, "port to listen on")
	replicaof := flag.String("replicaof", "", "Master host and port for replication")
	flag.Parse()

	role := "master"
	var masterHost string
	var masterPort int
	if *replicaof != "" {
		role = "slave"
		parts := strings.Split(*replicaof, " ")
		if len(parts) != 2 {
			fmt.Println("Invalid replicaof format. Expected: <MASTER_HOST> <MASTER_PORT>")
			os.Exit(1)
		}
		masterHost = parts[0]
		_, _ = fmt.Sscan(parts[1], &masterPort)
	}
	commands.SetServerRole(role)

	// 如果是副本，启动后台连接主节点的协程
	if role == "slave" {
		go func() {
			err := commands.InitiateReplication(masterHost, masterPort, *port)
			if err != nil {
				fmt.Println("InitiateReplication exist error: ", err)
				os.Exit(1)
			}
		}()
	}

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
