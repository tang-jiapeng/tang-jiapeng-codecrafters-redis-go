package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

// PingCommand 处理 PING 命令
type PingCommand struct {
}

func (c *PingCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("PING command takes no arguments")
	}
	return resp.EncodeSimpleString("PONG"), nil
}
