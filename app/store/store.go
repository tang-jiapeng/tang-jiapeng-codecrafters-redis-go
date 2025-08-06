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

// SetString 设置字符串键值对
func (s *Store) SetString(key, value string, expiresAt time.Time, hasExpiry bool) {
	s.Lock()
	defer s.Unlock()
	s.m[key] = StringEntry{
		Value:     value,
		ExpiresAt: expiresAt,
		HasExpiry: hasExpiry,
	}
	fmt.Printf("Store: SetString key=%s, value=%s, hasExpiry=%v, expiresAt=%v\n", key, value, hasExpiry, expiresAt)
}

// GetString 获取字符串值
func (s *Store) GetString(key string) (string, bool) {
	s.RLock()
	entry, exists := s.m[key]
	if !exists {
		fmt.Printf("Store: GetString key=%s, exists=false\n", key)
		return "", false
	}

	stringEntry, ok := entry.(StringEntry)
	if !ok {
		s.RUnlock()
		fmt.Printf("Store: GetString key=%s, invalid type (not a string)\n", key)
		return "", false
	}

	if stringEntry.HasExpiry && time.Now().After(stringEntry.ExpiresAt) {
		s.RUnlock()
		s.delete(key)
		fmt.Printf("Store: GetString key=%s, expired, deleted\n", key)
		return "", false
	}
	value := stringEntry.Value
	s.RUnlock()
	fmt.Printf("Store: GetString key=%s, value=%s\n", key, value)
	return value, true
}

// AppendList 追加元素到列表或创建新列表
func (s *Store) AppendList(key string, elements []string) (int, error) {
	s.Lock()
	defer s.Unlock()

	entry, exists := s.m[key]
	if !exists {
		list := ListEntry(elements)
		s.m[key] = list
		length := len(list)
		fmt.Printf("Store: AppendList key=%s, element=%s, new list=%v, length=%d\n", key, elements, list, length)
		return length, nil
	}

	list, ok := entry.(ListEntry)
	if !ok {
		fmt.Printf("Store: AppendList key=%s, invalid type (not a list)\n", key)
		return 0, fmt.Errorf("WRONGTYPE key is not a list")
	}
	list = append(list, elements...)
	s.m[key] = list
	length := len(list)
	fmt.Printf("Store: AppendList key=%s, element=%s, updated list=%v, length=%d\n", key, elements, list, length)
	return length, nil
}

// delete 删除键
func (s *Store) delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, key)
	fmt.Printf("Store: delete key=%s\n", key)
}
