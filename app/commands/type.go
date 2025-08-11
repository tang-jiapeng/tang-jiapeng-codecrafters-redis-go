package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type TypeCommand struct {
	stringOps store.StringOps
	listOps   store.ListOps
	streamOps store.StreamOps
}

func NewTypeCommand(s store.StringOps, l store.ListOps, ss store.StreamOps) *TypeCommand {
	return &TypeCommand{
		stringOps: s,
		listOps:   l,
		streamOps: ss,
	}
}

func (c *TypeCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("TYPE command requires exactly one argument")
	}
	key := args[0]
	if _, exists := c.stringOps.GetString(key); exists {
		return resp.EncodeSimpleString("string"), nil
	}
	if c.listOps.Exists(key) {
		return resp.EncodeSimpleString("list"), nil
	}
	if c.streamOps.Exists(key) {
		return resp.EncodeSimpleString("stream"), nil
	}
	return resp.EncodeSimpleString("none"), nil
}
