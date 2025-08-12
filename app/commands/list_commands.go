package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strconv"
	"time"
)

type LPushCommand struct {
	listOps store.ListOps
}

func NewLPushCommand(s store.ListOps) *LPushCommand {
	return &LPushCommand{
		listOps: s,
	}
}

func (c *LPushCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("RPUSH command requires at least two arguments")
	}
	key := args[0]
	elements := args[1:]

	length, err := c.listOps.PrependList(key, elements)
	if err != nil {
		return "", err
	}
	return resp.EncodeInteger(length), nil
}

type RPushCommand struct {
	listOps store.ListOps
}

func NewRPushCommand(s store.ListOps) *RPushCommand {
	return &RPushCommand{
		listOps: s,
	}
}

func (c *RPushCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("RPUSH command requires at least two arguments")
	}
	key := args[0]
	elements := args[1:]
	length, err := c.listOps.AppendList(key, elements)
	if err != nil {
		return "", err
	}
	return resp.EncodeInteger(length), nil
}

type LRangeCommand struct {
	listOps store.ListOps
}

func NewLRangeCommand(s store.ListOps) *LRangeCommand {
	return &LRangeCommand{
		listOps: s,
	}
}

func (c *LRangeCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("LRANGE command requires at least two arguments")
	}
	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return "", fmt.Errorf("invalid start index: %s", err.Error())
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return "", fmt.Errorf("invalid stop index: %s", err.Error())
	}
	elements, err := c.listOps.GetListRange(key, start, stop)
	if err != nil {
		return "", err
	}
	// 构建 RESP 数组
	respArray := make([]interface{}, len(elements))
	for i, elem := range elements {
		respArray[i] = elem
	}
	return resp.EncodeArray(respArray), nil
}

type LLenCommand struct {
	listOps store.ListOps
}

func NewLLenCommand(s store.ListOps) *LLenCommand {
	return &LLenCommand{
		listOps: s,
	}
}

func (c *LLenCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("LLEN command requires exactly one argument")
	}
	key := args[0]
	length, err := c.listOps.GetListLength(key)
	if err != nil {
		return "", err
	}
	return resp.EncodeInteger(length), nil
}

type LPopCommand struct {
	listOps store.ListOps
}

func NewLPopCommand(s store.ListOps) *LPopCommand {
	return &LPopCommand{
		listOps: s,
	}
}

func (c *LPopCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return "", fmt.Errorf("LPOP command requires one or two arguments")
	}
	key := args[0]
	count := 1
	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(args[1])
		if err != nil {
			return "", fmt.Errorf("invalid value not an integer or out of range: %s", args[1])
		}
		if count < 0 {
			return "", fmt.Errorf("count must be non-negative, got %d", count)
		}
	}
	elements, ok, err := c.listOps.LPopElement(key, count)
	if err != nil {
		return "", err
	}
	if !ok {
		return resp.EncodeNull(), nil
	}
	if count == 1 && len(elements) == 1 {
		// 单元素返回批量字符串
		return resp.EncodeBulkString(elements[0]), nil
	}
	// 多元素或 count>1 返回数组
	respArray := make([]interface{}, len(elements))
	for i, elem := range elements {
		respArray[i] = elem
	}
	return resp.EncodeArray(respArray), nil
}

type BLPopCommand struct {
	listOps store.ListOps
}

func NewBLPopCommand(s store.ListOps) *BLPopCommand {
	return &BLPopCommand{
		listOps: s,
	}
}

func (c *BLPopCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("BLPOP command requires two arguments")
	}
	key := args[0]
	timeoutSec, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return "", fmt.Errorf("invalid value timeout: %q", args[1])
	}
	timeout := time.Duration(timeoutSec * float64(time.Second))
	element, ok, err := c.listOps.BLPopElement(key, timeout)
	if err != nil {
		return "", err
	}
	if !ok {
		return resp.EncodeNull(), nil
	}
	// 构建 RESP 数组 [key, element]
	respArray := []interface{}{key, element}
	return resp.EncodeArray(respArray), nil
}
