package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type XAddCommand struct {
	streamOps store.StreamOps
}

func NewXAddCommand(ss store.StreamOps) *XAddCommand {
	return &XAddCommand{
		streamOps: ss,
	}
}

func (c *XAddCommand) Handle(args []string) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("XADD command requires at least three arguments")
	}
	streamKey := args[0]
	entryID := args[1]

	// 验证键值对参数
	fieldArgs := args[2:]
	if len(fieldArgs)%2 != 0 {
		return "", fmt.Errorf("wrong number of arguments for fields")
	}
	// 创建字段映射
	fields := make(map[string]string)
	for i := 0; i < len(fieldArgs); i += 2 {
		fields[fieldArgs[i]] = fieldArgs[i+1]
	}
	// 添加条目到流
	id, err := c.streamOps.AddEntry(streamKey, entryID, fields)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(id), id), nil
}
