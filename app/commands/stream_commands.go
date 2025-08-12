package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strconv"
	"strings"
	"time"
)

type XAddCommand struct {
	streamOps store.StreamOps
}

func NewXAddCommand(ss store.StreamOps) *XAddCommand {
	return &XAddCommand{
		streamOps: ss,
	}
}

func (c *XAddCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
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
	return resp.EncodeBulkString(id), nil
}

type XRangeCommand struct {
	streamOps store.StreamOps
}

func NewXRangeCommand(ss store.StreamOps) *XRangeCommand {
	return &XRangeCommand{
		streamOps: ss,
	}
}

func (c *XRangeCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if len(args) != 3 {
		return "", fmt.Errorf("XADD command requires exactly three arguments")
	}
	streamKey := args[0]
	startID, endID := args[1], args[2]
	entries, err := c.streamOps.GetRange(streamKey, startID, endID)
	if err != nil {
		return "", err
	}
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

type XReadCommand struct {
	streamOps store.StreamOps
}

func NewXReadCommand(ss store.StreamOps) *XReadCommand {
	return &XReadCommand{
		streamOps: ss,
	}
}

func (c *XReadCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
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
