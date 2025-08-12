package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"github.com/codecrafters-io/redis-starter-go/app/transaction"
	"strconv"
	"strings"
	"time"
)

type SetCommand struct {
	stringOps store.StringOps
}

func NewSetCommand(s store.StringOps) *SetCommand {
	return &SetCommand{
		stringOps: s,
	}
}

func (c *SetCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
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

	c.stringOps.SetString(key, value, expiresAt, hasExpiry)
	return resp.EncodeSimpleString("OK"), nil
}

type GetCommand struct {
	stringOps store.StringOps
}

func NewGetCommand(s store.StringOps) *GetCommand {
	return &GetCommand{
		stringOps: s,
	}
}

func (c *GetCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("GET command requires exactly one argument")
	}

	value, exists := c.stringOps.GetString(args[0])
	if !exists {
		return resp.EncodeNull(), nil
	}

	return resp.EncodeBulkString(value), nil
}

type IncrCommand struct {
	stringOps store.StringOps
}

func NewIncrCommand(s store.StringOps) *IncrCommand {
	return &IncrCommand{
		stringOps: s,
	}
}

func (c *IncrCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("INCR command requires at least one argument")
	}

	key := args[0]

	newValue, err := c.stringOps.Increment(key)
	if err != nil {
		return "", err
	}
	return resp.EncodeInteger(newValue), nil
}
