package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
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
	return "", nil
}
