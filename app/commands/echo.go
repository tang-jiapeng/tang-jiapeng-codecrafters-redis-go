package commands

import "fmt"

// EchoCommand 处理 ECHO 命令
type EchoCommand struct{}

func (c *EchoCommand) Handle(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("ECHO command requires exactly one argument")
	}
	fmt.Printf("ECHO command: arg=%s\n", args[0])
	return fmt.Sprintf("$%d\r\n%s\r\n", len(args[0]), args[0]), nil
}
