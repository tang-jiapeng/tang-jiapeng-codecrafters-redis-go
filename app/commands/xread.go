package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strings"
)

type XReadCommand struct {
	streamOps store.StreamOps
}

func NewXReadCommand(ss store.StreamOps) *XReadCommand {
	return &XReadCommand{
		streamOps: ss,
	}
}

func (c *XReadCommand) Handle(args []string) (string, error) {
	if len(args) < 3 || strings.ToUpper(args[0]) != "STREAMS" {
		return "", fmt.Errorf("XREAD command synatx error")
	}

	numStreams := (len(args) - 1) / 2
	if len(args)-1 != numStreams*2 {
		return "", fmt.Errorf("number of stream keys and IDs must match")
	}
	keys := args[1 : 1+numStreams]
	ids := args[1+numStreams:]

	result, err := c.streamOps.ReadStreams(keys, ids)
	if err != nil {
		return "", err
	}
	// 构建 RESP 数组
	respArray := make([]interface{}, 0, len(result))
	for _, key := range keys {
		entries, exists := result[key]
		if !exists || len(entries) == 0 {
			continue // 跳过空流
		}
		// 构建流的条目数组
		streamEntries := make([]interface{}, 0, len(entries))
		for _, entry := range entries {
			// 构建条目：[ID, [field1, value1, field2, value2, ...]]
			entryData := []interface{}{entry.ID}
			fields := make([]interface{}, 0, len(entry.Fields)*2)
			for field, value := range entry.Fields {
				fields = append(fields, field, value)
			}
			entryData = append(entryData, fields)
			streamEntries = append(streamEntries, entryData)
		}
		// 添加流结果：[key, [[id, [fields...]], ...]]
		respArray = append(respArray, []interface{}{key, streamEntries})
	}
	// 如果没有任何结果，返回nil（RESP null）
	if len(result) == 0 {
		return resp.EncodeNull(), nil
	}
	// 编码为RESP数组
	return resp.EncodeArray(respArray), nil
}
