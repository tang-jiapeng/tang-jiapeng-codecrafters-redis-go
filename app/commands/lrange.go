package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strconv"
)

type LRangeCommand struct {
	store *store.Store
}

func NewLRangeCommand(s *store.Store) *LRangeCommand {
	return &LRangeCommand{
		store: s,
	}
}

func (c *LRangeCommand) Handle(args []string) (string, error) {
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
	elements, err := c.store.GetListRange(key, start, stop)
	if err != nil {
		return "", err
	}
	resp := fmt.Sprintf("*%d\r\n", len(elements))
	for _, elem := range elements {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(elem), elem)
	}
	fmt.Printf("LRANGE key=%s, start=%d, stop=%d, result=%v\n", key, start, stop, elements)
	return resp, nil
}
