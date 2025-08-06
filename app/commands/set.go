package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"strconv"
	"strings"
	"time"
)

type SetCommand struct {
	store *store.Store
}

func NewSetCommand(s *store.Store) *SetCommand {
	return &SetCommand{store: s}
}

func (c *SetCommand) Handle(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("SET command requires at least two arguments")
	}
	if len(args) > 4 {
		return "", fmt.Errorf("SET command supports up to four arguments (key, value, PX, expiry)")
	}

	key := args[0]
	value := args[1]
	var expiresAt time.Time
	hasExpiry := false

	if len(args) == 4 {
		if strings.ToUpper(args[2]) != "PX" {
			return "", fmt.Errorf("invalid option: %s, expected PX", args[2])
		}
		expiryMs, err := strconv.Atoi(args[3])
		if err != nil {
			return "", fmt.Errorf("invalid PX value: %s", err.Error())
		}
		if expiryMs <= 0 {
			return "", fmt.Errorf("PX value must be positive")
		}
		expiresAt = time.Now().Add(time.Duration(expiryMs) * time.Millisecond)
		hasExpiry = true
	}

	c.store.SetString(key, value, expiresAt, hasExpiry)
	return "+OK\r\n", nil
}
