package store

import (
	"fmt"
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

// GetData 获取键对应的数据
func (s *Store) GetData(key string) (DataType, bool) {
	s.RLock()
	defer s.RUnlock()
	data, exists := s.m[key]
	if !exists {
		fmt.Printf("Store: GetData key=%s, exists=false\n", key)
		return nil, false
	}
	return data, true
}

// SetData 设置键值对
func (s *Store) SetData(key string, data DataType) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = data
	fmt.Printf("Store: SetData key=%s, type=%s\n", key, data.Type())
}

// Delete 删除键
func (s *Store) Delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
	fmt.Printf("Store: Delete key=%s\n", key)
}
