package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strconv"
	"strings"
	"time"
)

type XReadCommand struct {
	streamOps store.StreamOps
}

func NewXReadCommand(ss store.StreamOps) *XReadCommand {
	return &XReadCommand{
		streamOps: ss,
	}
}

func (c *XReadCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("XREAD command requires at least three arguments")
	}

	// 检查是否包含 BLOCK 选项
	var blockTimeout time.Duration
	var streamsIndex int
	if strings.ToUpper(args[0]) == "BLOCK" {
		if len(args) < 5 {
			return "", fmt.Errorf("XREAD with BLOCK requires at least five arguments")
		}
		timeoutMs, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid BLOCK timeout: %s", err.Error())
		}
		if timeoutMs < 0 {
			return "", fmt.Errorf("BLOCK timeout must be non-negative")
		}
		if strings.ToUpper(args[2]) != "STREAMS" {
			return "", fmt.Errorf("syntax error: expected STREAMS after BLOCK")
		}
		blockTimeout = time.Duration(timeoutMs) * time.Millisecond
		streamsIndex = 3
	} else if strings.ToUpper(args[0]) != "STREAMS" {
		return "", fmt.Errorf("syntax error")
	} else {
		streamsIndex = 1
	}

	// 分割流键和 ID
	numStreams := (len(args) - streamsIndex) / 2
	if len(args)-streamsIndex != numStreams*2 {
		return "", fmt.Errorf("number of stream keys and IDs must match")
	}
	keys := args[streamsIndex : streamsIndex+numStreams]
	ids := args[streamsIndex+numStreams:]

	// 调用阻塞或非阻塞查询
	var result map[string][]store.StreamEntry
	var err error
	if blockTimeout > 0 || streamsIndex == 3 {
		result, err = c.streamOps.ReadStreamsBlocking(keys, ids, blockTimeout)
	} else {
		result, err = c.streamOps.ReadStreams(keys, ids)
	}
	if err != nil {
		return "", err
	}
	// 如果没有任何结果，返回nil（RESP null）
	if result == nil || len(result) == 0 {
		return resp.EncodeNull(), nil
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
	// 编码为RESP数组
	return resp.EncodeArray(respArray), nil
}
