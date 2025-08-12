package commands

import (
	"fmt"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"strings"
)

type ReplconfCommand struct{}

func (c *ReplconfCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("REPLCONF requires at least one argument")
	}
	if strings.ToUpper(args[0]) == "GETACK" {
		return resp.EncodeArray([]interface{}{"REPLCONF", "ACK", "0"}), nil
	}
	return resp.EncodeSimpleString("OK"), nil
}

type PsyncCommand struct {
}

func (c *PsyncCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	msg := "FULLRESYNC " + GetMasterReplID() + " 0"
	return &RDBResponse{
		Message: resp.EncodeSimpleString(msg),
		RDBData: GetEmptyRDBData(),
	}, nil
}
