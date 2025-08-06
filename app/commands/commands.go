package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"net"
	"strings"
)

// CommandHandler 定义命令处理接口
type CommandHandler interface {
	Handle(args []string) (string, error)
}

// CommandRegistry 存储命令名称到处理器的映射
type CommandRegistry map[string]CommandHandler

// 全局存储，确保初始化
var storeInstance = store.NewStore()

// Commands 注册命令
var Commands = CommandRegistry{
	"PING":   &PingCommand{},
	"ECHO":   &EchoCommand{},
	"SET":    NewSetCommand(storeInstance),
	"GET":    NewGetCommand(storeInstance),
	"RPUSH":  NewRPushCommand(storeInstance),
	"LRANGE": NewLRangeCommand(storeInstance),
	"LPUSH":  NewLPushCommand(storeInstance),
	"LLEN":   NewLLenCommand(storeInstance),
}

// HandleConnection 处理客户端连接
func HandleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection: ", err.Error())
		}
	}(conn)

	reader := resp.NewRESPReader(conn)
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
