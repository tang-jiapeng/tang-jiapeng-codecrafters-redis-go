package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strconv"
	"time"
)

type BLPopCommand struct {
	listOps store.ListOps
}

func NewBLPopCommand(s store.ListOps) *BLPopCommand {
	return &BLPopCommand{
		listOps: s,
	}
}

func (c *BLPopCommand) Handle(args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("BLPOP command requires two arguments")
	}
	key := args[0]
	timeoutSecond, err := strconv.Atoi(args[1])
	if err != nil {
		return "", fmt.Errorf("invalid value timeout: %d", timeoutSecond)
	}
	timeout := time.Duration(timeoutSecond) * time.Second
	element, ok, err := c.listOps.BLPopElement(key, timeout)
	if err != nil {
		return "", err
	}
	if !ok {
		return "$-1\r\n", nil
	}
	return fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(element), element), nil
}
