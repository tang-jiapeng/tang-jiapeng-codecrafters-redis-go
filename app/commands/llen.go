package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type LLenCommand struct {
	listOps store.ListOps
}

func NewLLenCommand(s *store.Store) *LLenCommand {
	return &LLenCommand{
		listOps: store.NewListStore(s),
	}
}

func (c *LLenCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("LLEN command requires exactly one argument")
	}
	key := args[0]
	length, err := c.listOps.GetListLength(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(":%d\r\n", length), nil
}
