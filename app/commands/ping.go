package commands

import "fmt"

// PingCommand 处理 PING 命令
type PingCommand struct {
}

func (c *PingCommand) Handle(args []string) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("PING command takes no arguments")
	}
	fmt.Println("PING command executed")
	return "+PONG\r\n", nil
}
