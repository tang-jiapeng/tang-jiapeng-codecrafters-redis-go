package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type RpushCommand struct {
	store *store.Store
}

func NewRpushCommand(s *store.Store) *RpushCommand {
	return &RpushCommand{
		store: s,
	}
}

func (c *RpushCommand) Handle(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("RPUSH command requires at least two arguments")
	}
	key := args[0]
	elements := args[1:]
	length, err := c.store.AppendList(key, elements)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(":%d\r\n", length), nil
}
