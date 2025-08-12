package commands

import "github.com/codecrafters-io/redis-starter-go/app/resp"

type PsyncCommand struct {
}

func (c *PsyncCommand) Handle(ctx *ConnectionContext, args []string) (string, error) {
	response := "FULLRESYNC " + MasterReplID + " 0"
	return resp.EncodeSimpleString(response), nil
}
