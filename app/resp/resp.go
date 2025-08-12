package resp

import (
	"bufio"
	"fmt"
	"io"
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

// Value 表示 RESP 协议中的各种数据类型
type Value struct {
	Type    string  // 数据类型: "simple_string", "error", "integer", "bulk_string", "array"
	String  string  // 用于 simple_string, error, bulk_string
	Integer int     // 用于 integer
	Array   []Value // 用于 array
}

// FormatForDisplay 格式化Value用于调试输出
func (v Value) FormatForDisplay() string {
	switch v.Type {
	case "simple_string":
		return fmt.Sprintf("(simple_string) %s", v.String)
	case "error":
		return fmt.Sprintf("(error) %s", v.String)
	case "integer":
		return fmt.Sprintf("(integer) %d", v.Integer)
	case "bulk_string":
		return fmt.Sprintf("(bulk_string) %s", v.String)
	case "array":
		items := make([]string, len(v.Array))
		for i, item := range v.Array {
			items[i] = item.FormatForDisplay()
		}
		return fmt.Sprintf("(array) [%s]", strings.Join(items, ", "))
	default:
		return "(unknown)"
	}
}

// Read 通用RESP解析方法，可以处理所有RESP数据类型
func (r *RESPReader) Read() (Value, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return Value{}, err
	}
	line = strings.TrimSuffix(line, "\r\n")
	if len(line) == 0 {
		return Value{}, fmt.Errorf("empty response")
	}
	switch line[0] {
	case '+':
		// 简单字符串: +OK\r\n
		return Value{Type: "simple_string", String: line[1:]}, nil
	case '-':
		// 错误: -ERR something\r\n
		return Value{Type: "error", String: line[1:]}, nil
	case ':':
		// 整数: :1000\r\n
		intValue, err := strconv.Atoi(line[1:])
		if err != nil {
			return Value{}, nil
		}
		return Value{Type: "integer", Integer: intValue}, nil
	case '$':
		// 批量字符串: $5\r\nhello\r\n
		strLen, err := strconv.Atoi(line[1:])
		if err != nil {
			return Value{}, err
		}
		if strLen == -1 {
			// 空批量字符串 (Redis NULL)
			return Value{Type: "bulk_string", String: ""}, nil
		}
		// 读取指定长度的字符串
		data := make([]byte, strLen)
		_, err = io.ReadFull(r.reader, data)
		if err != nil {
			return Value{}, nil
		}
		// 消耗尾部的 \r\n
		_, err = r.reader.Discard(2)
		if err != nil {
			return Value{}, err
		}
		return Value{Type: "bulk_string", String: string(data)}, nil
	case '*':
		// 数组: *2\r\n$5\r\nhello\r\n$5\r\nworld\r\n
		arrayLen, err := strconv.Atoi(line[1:])
		if err != nil {
			return Value{}, err
		}
		// 处理空数组
		if arrayLen == -1 {
			return Value{Type: "array", Array: nil}, nil
		}
		array := make([]Value, arrayLen)
		for i := 0; i < arrayLen; i++ {
			val, err := r.Read()
			if err != nil {
				return Value{}, err
			}
			array[i] = val
		}
		return Value{Type: "array", Array: array}, nil
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %s", string(line[0]))
	}
}

// ReadCommand 读取并解析一个 RESP 命令
func (r *RESPReader) ReadCommand() ([]string, error) {
	value, err := r.Read()
	if err != nil {
		return nil, err
	}

	if value.Type != "array" {
		return nil, fmt.Errorf("expected RESP array, got %s", value.Type)
	}

	args := make([]string, len(value.Array))
	for i, item := range value.Array {
		if item.Type != "bulk_string" {
			return nil, fmt.Errorf("expected bulk string in array, got %s", item.Type)
		}
		args[i] = item.String
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
	return fmt.Sprintf("-ERR %s\r\n", msg)
}
