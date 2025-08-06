package main

import (
	"fmt"
	"net"
	"strings"
)

// CommandHandler 定义命令处理接口
type CommandHandler interface {
	Handle(args []string) (string, error)
}

type CommandRegistry map[string]CommandHandler

var Commands = CommandRegistry{
	"PING": &PingCommand{},
	"ECHO": &EchoCommand{},
}

// PingCommand 处理 PING 命令
type PingCommand struct{}

func (c *PingCommand) Handle(args []string) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("PING command tasks no arguments")
	}
	return "+PONG\r\n", nil
}

// EchoCommand 处理 ECHO 命令
type EchoCommand struct{}

func (c *EchoCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("ECHO command requires exactly one argument")
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(args[0]), args[0]), nil
}

// handleConnection 处理客户端连接
func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection: ", err.Error())
		}
	}(conn)

	reader := NewRESPReader(conn)
	for {
		args, err := reader.ReadCommand()
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Connection closed")
				break
			}
			fmt.Println("Error parsing RESP: ", err.Error())
			break
		}

		if len(args) == 0 {
			fmt.Println("Empty command received")
			continue
		}

		commandName := strings.ToUpper(args[0])
		handler, exists := Commands[commandName]
		if !exists {
			fmt.Println("Unknown command: ", commandName)
			conn.Write([]byte("-ERR unknown command\r\n"))
			continue
		}
		response, err := handler.Handle(args[1:])
		if err != nil {
			fmt.Println("Error handling command: ", err.Error())
			conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", err.Error())))
			continue
		}
		conn.Write([]byte(response))
	}
}
