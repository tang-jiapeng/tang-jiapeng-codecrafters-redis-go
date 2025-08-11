package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type MultiCommand struct{}

func (c *MultiCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	// 如果已经在事务模式，返回错误
	if ctx.InTransaction {
		return "", fmt.Errorf("MULTI calls can not be nested")
	}
	ctx.InTransaction = true
	ctx.QueuedCommands = make([][]string, 0) // 清空事务队列
	return resp.EncodeSimpleString("OK"), nil
}
