package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"strings"
)

type MultiCommand struct{}

func (c *MultiCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	// 如果已经在事务模式，返回错误
	if ctx.InTransaction {
		return "", fmt.Errorf("MULTI calls can not be nested")
	}
	ctx.InTransaction = true
	ctx.QueuedCommands = make([][]string, 0) // 清空事务队列
	return resp.EncodeSimpleString("OK"), nil
}

type ExecCommand struct {
}

func (c *ExecCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if !ctx.InTransaction {
		// 实现未进入事务时的错误提示
		return "", fmt.Errorf("EXEC without MULTI")
	}
	// 执行空事务
	if len(ctx.QueuedCommands) == 0 {
		ctx.InTransaction = false
		return resp.EncodeArray(nil), nil
	}

	results := make([]interface{}, 0, len(ctx.QueuedCommands))
	for _, cmdArgs := range ctx.QueuedCommands {
		commandName := strings.ToUpper(cmdArgs[0])
		handler, exists := Commands[commandName]
		if !exists {
			// 不存在的命令 → 返回错误响应
			results = append(results, resp.EncodeError("unknown command '"+cmdArgs[0]+"'"))
			continue
		}

		respValue, err := handler.Handle(ctx, cmdArgs[1:])
		if err != nil {
			results = append(results, resp.EncodeError(err.Error()))
		} else {
			switch v := respValue.(type) {
			case string:
				results = append(results, v)
			case *RDBResponse:
				// 事务中不支持RDB响应
				results = append(results, resp.EncodeError("RDB response not allowed in transaction"))
			default:
				// 处理其他类型
				results = append(results, resp.EncodeError("unsupported response type"))
			}
		}
	}
	ctx.InTransaction = false
	ctx.QueuedCommands = nil

	return resp.EncodeArray(results), nil
}

type DiscardCommand struct {
}

func (c *DiscardCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if !ctx.InTransaction {
		return "", fmt.Errorf("DISCARD without MULTI")
	}
	// 丢弃队列，退出事务
	ctx.QueuedCommands = nil
	ctx.InTransaction = false
	return resp.EncodeSimpleString("OK"), nil
}
