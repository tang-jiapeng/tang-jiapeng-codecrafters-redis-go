package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strconv"
)

type LPopCommand struct {
	listOps store.ListOps
}

func NewLPopCommand(s store.ListOps) *LPopCommand {
	return &LPopCommand{
		listOps: s,
	}
}

func (c *LPopCommand) Handle(args []string) (string, error) {
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
