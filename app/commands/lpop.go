package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type LPopCommand struct {
	listOps store.ListOps
}

func NewLPopCommand(s *store.Store) *LPopCommand {
	return &LPopCommand{
		listOps: store.NewListStore(s),
	}
}

func (c *LPopCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("LPOP command requires exactly one argument")
	}
	key := args[0]
	element, ok, err := c.listOps.PopLElement(key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "$-1\r\n", nil
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(element), element), nil
}
