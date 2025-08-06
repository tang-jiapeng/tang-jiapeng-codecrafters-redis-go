package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// RESPReader 封装 RESP 解析逻辑
type RESPReader struct {
	reader *bufio.Reader
}

func NewRESPReader(conn net.Conn) *RESPReader {
	return &RESPReader{
		reader: bufio.NewReader(conn),
	}
}

// ReadCommand 读取并解析一个 RESP 命令
func (r *RESPReader) ReadCommand() ([]string, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(line, "*") {
		return nil, fmt.Errorf("invalid RESP array")
	}
	numElements, err := strconv.Atoi(strings.TrimSuffix(line[1:], "\r\n"))
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %s", err.Error())
	}
	args := make([]string, 0, numElements)
	for i := 0; i < numElements; i++ {
		line, err := r.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading bulk string length: %s", err.Error())
		}
		if !strings.HasPrefix(line, "$") {
			return nil, fmt.Errorf("invalid bulk string")
		}
		strLen, err := strconv.Atoi(strings.TrimSuffix(line[1:], "\r\n"))
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length: %s", err.Error())
		}

		line, err = r.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading bulk string content: %s", err.Error())
		}
		content := strings.TrimSuffix(line, "\r\n")
		if len(content) != strLen {
			return nil, fmt.Errorf("bulk string length mismatch")
		}
		args = append(args, content)
	}
	return args, nil
}
