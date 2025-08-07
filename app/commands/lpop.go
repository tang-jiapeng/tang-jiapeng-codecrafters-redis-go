package commands

import (
	"fmt"
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
	elements, ok, err := c.listOps.PopLElement(key, count)
	if err != nil {
		return "", err
	}
	if !ok {
		return "$-1\r\n", nil
	}
	if count == 1 && len(elements) == 1 {
		// 单元素返回批量字符串
		return fmt.Sprintf("$%d\r\n%s\r\n", len(elements[0]), elements[0]), nil
	}
	// 多元素或 count>1 返回数组
	resp := fmt.Sprintf("*%d\r\n", len(elements))
	for _, elem := range elements {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(elem), elem)
	}
	fmt.Printf("LPOP key=%s, count=%d, popped=%v\n", key, count, elements)
	return resp, nil
}
