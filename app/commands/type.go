package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type TypeCommand struct {
	stringOps store.StringOps
	listOps   store.ListOps
}

func NewTypeCommand(s store.StringOps, l store.ListOps) *TypeCommand {
	return &TypeCommand{
		stringOps: s,
		listOps:   l,
	}
}

func (c *TypeCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("TYPE command requires exactly one argument")
	}
	key := args[0]
	if _, exists := c.stringOps.GetString(key); exists {
		return "+string\r\n", nil
	}

	if c.listOps.Exists(key) {
		return "+list\r\n", nil
	}
	return "+none\r\n", nil
}
