package store

import "time"

// StringEntry 表示字符串键值对及其过期时间
type StringEntry struct {
	Value     string
	ExpiresAt time.Time
	HasExpiry bool
}

// ListEntry 表示列表类型数据
type ListEntry []string

func (StringEntry) Type() string {
	return "string"
}

func (ListEntry) Type() string {
	return "list"
}
