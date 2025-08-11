package commands

import (
	"fmt"
)

type ExecCommand struct {
}

func (c *ExecCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	if !ctx.InTransaction {
		// 实现未进入事务时的错误提示
		return "", fmt.Errorf("EXEC without MULTI")
	}

	return "", nil
}
