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
	Handle(ctx *ConnectionContext, args []string) (string, error)
}

// CommandRegistry 存储命令名称到处理器的映射
type CommandRegistry map[string]CommandHandler

// 初始化独立的 Store 实例
var stringStore = store.NewStringStore()
var listStore = store.NewListStore()
var streamStore = store.NewStreamStore()

// Commands 注册命令
var Commands = CommandRegistry{
	"PING":     &PingCommand{},
	"ECHO":     &EchoCommand{},
	"COMMAND":  &NoOpCommand{}, // 空实现
	"REPLCONF": &NoOpCommand{},
	"PSYNC":    &PsyncCommand{},
	"INFO":     &InfoCommand{},
	"MULTI":    &MultiCommand{},
	"EXEC":     &ExecCommand{},
	"DISCARD":  &DiscardCommand{},
	"SET":      NewSetCommand(stringStore),
	"GET":      NewGetCommand(stringStore),
	"INCR":     NewIncrCommand(stringStore),
	"RPUSH":    NewRPushCommand(listStore),
	"LRANGE":   NewLRangeCommand(listStore),
	"LPUSH":    NewLPushCommand(listStore),
	"LLEN":     NewLLenCommand(listStore),
	"LPOP":     NewLPopCommand(listStore),
	"BLPOP":    NewBLPopCommand(listStore),
	"TYPE":     NewTypeCommand(stringStore, listStore, streamStore),
	"XADD":     NewXAddCommand(streamStore),
	"XRANGE":   NewXRangeCommand(streamStore),
	"XREAD":    NewXReadCommand(streamStore),
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
	// 每个连接单独一个事务上下文
	connCtx := NewConnectionContext()

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
			continue
		}

		commandName := strings.ToUpper(args[0])
		handler, exists := Commands[commandName]
		if !exists {
			respErr := resp.EncodeError("unknown command '" + commandName + "'")
			conn.Write([]byte(respErr))
			continue
		}

		// 事务模式下且命令不是 MULTI/EXEC/DISCARD就排队
		if connCtx.InTransaction && (commandName != "MULTI" && commandName != "EXEC" && commandName != "DISCARD") {
			connCtx.QueuedCommands = append(connCtx.QueuedCommands, args)
			conn.Write([]byte("+QUEUED\r\n"))
			continue
		}

		// 正常执行
		response, err := handler.Handle(connCtx, args[1:])
		if err != nil {
			conn.Write([]byte(resp.EncodeError(err.Error())))
			continue
		}
		conn.Write([]byte(response))
	}
}

// 空实现
type NoOpCommand struct{}

func (c *NoOpCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	return resp.EncodeSimpleString("OK"), nil
}
