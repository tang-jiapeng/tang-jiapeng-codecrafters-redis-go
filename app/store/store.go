package store

import (
	"sync"
	"time"
)

// Store 封装线程安全的键值存储
type Store struct {
	sync.RWMutex
	m map[string]DataType
}

func NewStore() *Store {
	return &Store{
		m: make(map[string]DataType),
	}
}

// DataType 接口定义数据类型操作
type DataType interface {
	Type() string
}

// StringOps 定义字符串操作接口
type StringOps interface {
	SetString(key, value string, expiresAt time.Time, hasExpiry bool)
	GetString(key string) (string, bool)
}

// ListOps 定义列表操作接口
type ListOps interface {
	AppendList(key string, elements []string) (int, error)
	PrependList(key string, elements []string) (int, error)
	GetListRange(key string, start, stop int) ([]string, error)
	GetListLength(key string) (int, error)
	PopLElement(key string) (string, bool, error)
}
