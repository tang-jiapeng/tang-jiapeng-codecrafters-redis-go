package resp

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

// EncodeArray 编码RESP数组
func EncodeArray(data []interface{}) string {
	// 返回空数组 *0\r\n，而不是 null
	result := fmt.Sprintf("*%d\r\n", len(data))
	for _, item := range data {
		switch v := item.(type) {
		case string:
			result += EncodeBulkString(v)
		case []interface{}:
			result += EncodeArray(v)
		default:
			result += EncodeBulkString(fmt.Sprintf("%v", v))
		}
	}
	return result
}

// EncodeInteger 编码 RESP 整数
func EncodeInteger(i int) string {
	return fmt.Sprintf(":%d\r\n", i)
}

// EncodeBulkString 编码RESP批量字符串
func EncodeBulkString(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

// EncodeNull 编码RESP null
func EncodeNull() string {
	return "$-1\r\n"
}

// EncodeSimpleString 编码 RESP 简单字符串
func EncodeSimpleString(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}

// EncodeError 编码 RESP 错误
func EncodeError(msg string) string {
	return fmt.Sprintf("-%s\r\n", msg)
}

// EncodeArrayRaw 用于直接拼接已经编码好的RESP元素（字符串）
// 用于事务中直接返回已编码的命令执行结果
func EncodeArrayRaw(elements []string) string {
	result := fmt.Sprintf("*%d\r\n", len(elements))
	for _, e := range elements {
		result += e
	}
	return result
}
