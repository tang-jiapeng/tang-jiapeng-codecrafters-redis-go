package store

import "time"

// DataType 定义支持的数据类型
type DataType interface {
	isDataType()
}

// StringEntry 表示字符串键值对及其过期时间
type StringEntry struct {
	Value     string
	ExpiresAt time.Time
	HasExpiry bool
}

type ListEntry []string

// 实现 DataType 接口
func (StringEntry) isDataType() {}
func (ListEntry) isDataType()   {}
