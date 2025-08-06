package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CommandHandler 定义命令处理接口
type CommandHandler interface {
	Handle(args []string) (string, error)
}

type CommandRegistry map[string]CommandHandler

// StoreEntry 表示键值对及其过期时间
type StoreEntry struct {
	value     string
	expiresAt time.Time
	hasExpiry bool
}

// Store 封装线程安全的键值存储
type Store struct {
	sync.RWMutex
	m map[string]StoreEntry
}

// 全局存储
var store = Store{
	m: make(map[string]StoreEntry),
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
	if len(args) < 2 {
		return "", fmt.Errorf("SET command requires at least two arguments")
	}
	if len(args) > 4 {
		return "", fmt.Errorf("SET command supports up to four arguments (key, value, PX, expiry)")
	}

	key := args[0]
	value := args[1]
	var expiresAt time.Time
	hasExpiry := false

	if len(args) == 4 {
		if strings.ToUpper(args[2]) != "PX" {
			return "", fmt.Errorf("invalid option: %s, expected PX", args[2])
		}
		expiryMs, err := strconv.Atoi(args[3])
		if err != nil {
			return "", fmt.Errorf("invalid PX value: %s", err.Error())
		}
		if expiryMs < 0 {
			return "", fmt.Errorf("PX value must be positive")
		}
		expiresAt = time.Now().Add(time.Duration(expiryMs) * time.Millisecond)
		hasExpiry = true
	}

	store.Lock()
	store.m[key] = StoreEntry{
		value:     value,
		expiresAt: expiresAt,
		hasExpiry: hasExpiry,
	}
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
	entry, exists := store.m[args[0]]
	store.RUnlock()
	if !exists {
		fmt.Printf("GET key=%s, exists=false, store=%v\n", args[0], store.m)
		return "$-1\r\n", nil
	}
	if entry.hasExpiry && time.Now().After(entry.expiresAt) {
		store.Lock()
		delete(store.m, args[0])
		store.Unlock()
		fmt.Printf("GET key=%s, expired, deleted, store=%v\n", args[0], store.m)
		return "$-1\r\n", nil
	}
	fmt.Printf("GET key=%s, exists=true, value=%s, store=%v\n", args[0], entry.value, store.m)
	return fmt.Sprintf("$%d\r\n%s\r\n", len(entry.value), entry.value), nil
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
