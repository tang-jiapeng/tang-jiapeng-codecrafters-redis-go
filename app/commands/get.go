package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type GetCommand struct {
	stringOps store.StringOps
}

func NewGetCommand(s store.StringOps) *GetCommand {
	return &GetCommand{
		stringOps: s,
	}
}

func (c *GetCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("GET command requires exactly one argument")
	}

	value, exists := c.stringOps.GetString(args[0])
	if !exists {
		return "$-1\r\n", nil
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value), nil
}
