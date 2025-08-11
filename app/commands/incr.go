package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type IncrCommand struct {
	stringOps store.StringOps
}

func NewIncrCommand(s store.StringOps) *IncrCommand {
	return &IncrCommand{
		stringOps: s,
	}
}

func (c *IncrCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("INCR command requires at least one argument")
	}

	key := args[0]

	newValue, err := c.stringOps.Increment(key)
	if err != nil {
		return "", err
	}
	return resp.EncodeInteger(newValue), nil
}
