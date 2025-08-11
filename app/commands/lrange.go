package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strconv"
)

type LRangeCommand struct {
	listOps store.ListOps
}

func NewLRangeCommand(s store.ListOps) *LRangeCommand {
	return &LRangeCommand{
		listOps: s,
	}
}

func (c *LRangeCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
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
