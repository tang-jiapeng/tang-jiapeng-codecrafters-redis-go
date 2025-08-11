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

	// 构建 RESP 数组
	respArray := make([]interface{}, 0, len(entries))
	for _, entry := range entries {
		// 每个条目是一个数组：[ID, [field1, value1, field2, value2, ...]]
		entryData := make([]interface{}, 2)
		entryData[0] = entry.ID // 条目 ID
		// 构建字段数组
		fields := make([]interface{}, 0, len(entry.Fields)*2)
		for key, value := range entry.Fields {
			fields = append(fields, key, value)
		}
		entryData[1] = fields // 字段列表
		respArray = append(respArray, entryData)
	}

	// 编码为 RESP 数组
	return resp.EncodeArray(respArray), nil
}
