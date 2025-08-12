package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"github.com/codecrafters-io/redis-starter-go/app/transaction"
	"strings"
)

// NoOpCommand 空实现
type NoOpCommand struct{}

func (c *NoOpCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	return resp.EncodeSimpleString("OK"), nil
}

// PingCommand 处理 PING 命令
type PingCommand struct {
}

func (c *PingCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("PING command takes no arguments")
	}
	return resp.EncodeSimpleString("PONG"), nil
}

// EchoCommand 处理 ECHO 命令
type EchoCommand struct{}

func (c *EchoCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("ECHO command requires exactly one argument")
	}
	return resp.EncodeBulkString(args[0]), nil
}

type InfoCommand struct{}

func (c *InfoCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	section := ""
	if len(args) > 0 {
		section = strings.ToLower(args[0])
	}
	if section == "replication" {
		// 返回 replication 信息
		info := fmt.Sprintf("role:%s\r\n", replication.GetServerRole()) +
			fmt.Sprintf("master_replid:%s\r\n", replication.GetMasterReplID()) +
			fmt.Sprintf("master_repl_offset:%d\r\n", replication.GetMasterReplOffset())
		return resp.EncodeBulkString(info), nil
	}
	// 默认返回空或其他 sections
	return "", nil
}

type TypeCommand struct {
	stringOps store.StringOps
	listOps   store.ListOps
	streamOps store.StreamOps
}

func NewTypeCommand(s store.StringOps, l store.ListOps, ss store.StreamOps) *TypeCommand {
	return &TypeCommand{
		stringOps: s,
		listOps:   l,
		streamOps: ss,
	}
}

func (c *TypeCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("TYPE command requires exactly one argument")
	}
	key := args[0]
	if _, exists := c.stringOps.GetString(key); exists {
		return resp.EncodeSimpleString("string"), nil
	}
	if c.listOps.Exists(key) {
		return resp.EncodeSimpleString("list"), nil
	}
	if c.streamOps.Exists(key) {
		return resp.EncodeSimpleString("stream"), nil
	}
	return resp.EncodeSimpleString("none"), nil
}
