package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"strings"
)

type InfoCommand struct {
}

func (c *InfoCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	section := ""
	if len(args) > 0 {
		section = strings.ToLower(args[0])
	}
	if section == "replication" {
		// 返回 replication 信息
		info := fmt.Sprintf("role:%s\r\n", GetServerRole()) +
			fmt.Sprintf("master_replid:%s\r\n", MasterReplID) +
			fmt.Sprintf("master_repl_offset:%d\r\n", MasterReplOffset)
		return resp.EncodeBulkString(info), nil
	}
	// 默认返回空或其他 sections
	return "", nil
}
