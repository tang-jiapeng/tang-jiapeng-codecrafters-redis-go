package commands

import (
	"fmt"
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
		resp := fmt.Sprintf(
			"# Replication\r\nrole:master\r\nmaster_replid:%s\r\nmaster_repl_offset:%d\r\n",
			MasterReplID, MasterReplOffset,
		)
		return resp, nil
	}
	// 默认返回空或其他 sections
	return "", nil
}
