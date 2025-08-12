package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"strings"
)

type InfoCommand struct {
}

func (c *InfoCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	// 当前支持 "INFO replication"
	if len(args) > 0 && strings.ToLower(args[0]) == "replication" {
		return resp.EncodeBulkString("role:master"), nil
	}
	// 其他 INFO 情况暂不处理
	return resp.EncodeBulkString("role:master"), nil
}
