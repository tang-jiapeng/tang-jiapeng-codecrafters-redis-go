package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strings"
)

type XRangeCommand struct {
	streamOps store.StreamOps
}

func NewXRangeCommand(ss store.StreamOps) *XRangeCommand {
	return &XRangeCommand{
		streamOps: ss,
	}
}

func (c *XRangeCommand) Handle(args []string) (string, error) {
	if len(args) != 3 {
		return "", fmt.Errorf("XADD command requires exactly three arguments")
	}
	streamKey := args[0]
	startID, endID := args[1], args[2]
	entries, err := c.streamOps.GetRange(streamKey, startID, endID)
	if err != nil {
		return "", err
	}
	var respBuilder strings.Builder

	respBuilder.WriteString(fmt.Sprintf("*%d\r\n", len(entries)))

	for _, entry := range entries {
		// 每个条目是一个包含两个元素的数组：[ID, 字段列表]
		respBuilder.WriteString("*2\r\n")
		// 条目 ID (RESP 批量字符串)
		respBuilder.WriteString(resp.BulkString(entry.ID))
		// 字段列表 (RESP 数组)
		fieldCount := len(entry.Fields) * 2
		respBuilder.WriteString(fmt.Sprintf("*%d\r\n", fieldCount))

		for key, value := range entry.Fields {
			respBuilder.WriteString(resp.BulkString(key))
			respBuilder.WriteString(resp.BulkString(value))
		}
	}
	return respBuilder.String(), nil
}
