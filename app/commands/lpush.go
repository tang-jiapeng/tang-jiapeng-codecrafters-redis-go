package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type LPushCommand struct {
	listStore *store.ListStore
}

func NewLPushCommand(s *store.Store) *LPushCommand {
	return &LPushCommand{
		listStore: store.NewListStore(s),
	}
}

func (c *LPushCommand) Handle(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("RPUSH command requires at least two arguments")
	}
	key := args[0]
	elements := args[1:]

	length, err := c.listStore.PrependList(key, elements)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(":%d\r\n", length), nil
}
