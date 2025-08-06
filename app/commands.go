package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// CommandHandler 定义命令处理接口
type CommandHandler interface {
	Handle(args []string) (string, error)
}

type CommandRegistry map[string]CommandHandler

// Store 封装线程安全的键值存储
type Store struct {
	sync.RWMutex
	m map[string]string
}

// 全局存储
var store = Store{
	m: make(map[string]string),
}

var Commands = CommandRegistry{
	"PING": &PingCommand{},
	"ECHO": &EchoCommand{},
	"SET":  &SetCommand{},
	"GET":  &GetCommand{},
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

// SetCommand 处理 SET 命令
type SetCommand struct{}

func (c *SetCommand) Handle(args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("SET command requires exactly two arguments")
	}
	store.Lock()
	store.m[args[0]] = args[1]
	store.Unlock()
	fmt.Printf("SET key=%s, value=%s, store=%v\n", args[0], args[1], store.m)
	return "+OK\r\n", nil
}

// GetCommand 处理 GET 命令
type GetCommand struct{}

func (c *GetCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("GET command requires exactly one argument")
	}
	store.RLock()
	value, exists := store.m[args[0]]
	store.RUnlock()
	fmt.Printf("GET key=%s, exists=%v, value=%s, store=%v\n", args[0], exists, value, store.m)
	if !exists {
		return "$-1\r\n", nil
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value), nil
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
