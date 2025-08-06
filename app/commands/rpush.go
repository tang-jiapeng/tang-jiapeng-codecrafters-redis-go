package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type RPushCommand struct {
	listStore *store.ListStore
}

func NewRPushCommand(s *store.Store) *RPushCommand {
	return &RPushCommand{
		listStore: store.NewListStore(s),
	}
}

func (c *RPushCommand) Handle(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("RPUSH command requires at least two arguments")
	}
	key := args[0]
	elements := args[1:]
	length, err := c.listStore.AppendList(key, elements)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(":%d\r\n", length), nil
}
