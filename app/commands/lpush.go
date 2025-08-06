package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type LPushCommand struct {
	store *store.Store
}

func NewLPushCommand(s *store.Store) *LPushCommand {
	return &LPushCommand{
		store: s,
	}
}

func (c *LPushCommand) Handle(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("RPUSH command requires at least two arguments")
	}
	key := args[0]
	elements := args[1:]

	length, err := c.store.PrependList(key, elements)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(":%d\r\n", length), nil
}
