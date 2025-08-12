package commands

import "github.com/codecrafters-io/redis-starter-go/app/resp"

type ReplconfCommand struct{}

func (c *ReplconfCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	return resp.EncodeSimpleString("OK"), nil
}

type PsyncCommand struct {
}

func (c *PsyncCommand) Handle(ctx *ConnectionContext, args []string) (interface{}, error) {
	msg := "FULLRESYNC " + MasterReplID + " 0"
	return &RDBResponse{
		Message: resp.EncodeSimpleString(msg),
		RDBData: emptyRDBData,
	}, nil
}
