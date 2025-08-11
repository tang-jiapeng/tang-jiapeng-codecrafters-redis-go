package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

// EchoCommand 处理 ECHO 命令
type EchoCommand struct{}

func (c *EchoCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("ECHO command requires exactly one argument")
	}
	return resp.EncodeBulkString(args[0]), nil
}
