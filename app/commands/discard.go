package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

type DiscardCommand struct {
}

func (c *DiscardCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	if !ctx.InTransaction {
		return "", fmt.Errorf("DISCARD without MULTI")
	}
	// 丢弃队列，退出事务
	ctx.QueuedCommands = nil
	ctx.InTransaction = false
	return resp.EncodeSimpleString("OK"), nil
}
