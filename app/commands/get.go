package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type GetCommand struct {
	store *store.Store
}

func NewGetCommand(s *store.Store) *GetCommand {
	return &GetCommand{store: s}
}

func (c *GetCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("GET command requires exactly one argument")
	}

	value, exists := c.store.GetString(args[0])
	if !exists {
		return "$-1\r\n", nil
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value), nil
}
