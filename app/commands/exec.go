package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"strings"
)

type ExecCommand struct {
}

func (c *ExecCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	if !ctx.InTransaction {
		// 实现未进入事务时的错误提示
		return "", fmt.Errorf("EXEC without MULTI")
	}
	// 执行空事务
	if len(ctx.QueuedCommands) == 0 {
		ctx.InTransaction = false
		return resp.EncodeArray(nil), nil
	}

	results := make([]string, 0, len(ctx.QueuedCommands))
	for _, cmdArgs := range ctx.QueuedCommands {
		commandName := strings.ToUpper(cmdArgs[0])
		handler, exists := Commands[commandName]
		if !exists {
			// 不存在的命令 → 返回错误响应
			results = append(results, resp.EncodeError("unknown command '"+cmdArgs[0]+"'"))
			continue
		}

		respStr, err := handler.Handle(ctx, cmdArgs[1:])
		if err != nil {
			results = append(results, resp.EncodeError(err.Error()))
		} else {
			results = append(results, respStr)
		}
	}
	ctx.InTransaction = false
	ctx.QueuedCommands = nil

	return resp.EncodeArrayRaw(results), nil
}
