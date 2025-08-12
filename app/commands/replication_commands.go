package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/transaction"
)

type ReplconfCommand struct{}

func (c *ReplconfCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	return resp.EncodeSimpleString("OK"), nil
}

type PsyncCommand struct {
}

func (c *PsyncCommand) Handle(ctx *transaction.ConnectionContext, args []string) (interface{}, error) {
	msg := "FULLRESYNC " + replication.GetMasterReplID() + " 0"
	return &RDBResponse{
		Message: resp.EncodeSimpleString(msg),
		RDBData: replication.GetEmptyRDBData(),
	}, nil
}
