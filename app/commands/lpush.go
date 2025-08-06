package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type LPushCommand struct {
	listOps store.ListOps
}

func NewLPushCommand(s *store.Store) *LPushCommand {
	return &LPushCommand{
		listOps: store.NewListStore(s),
	}
}

func (c *LPushCommand) Handle(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("RPUSH command requires at least two arguments")
	}
	key := args[0]
	elements := args[1:]

	length, err := c.listOps.PrependList(key, elements)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(":%d\r\n", length), nil
}
