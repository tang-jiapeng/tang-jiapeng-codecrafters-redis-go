package store

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

// StringOps 定义字符串操作接口
type StringOps interface {
	SetString(key, value string, expiresAt time.Time, hasExpiry bool)
	GetString(key string) (string, bool)
	Increment(key string) (int, error)
}

// StringEntry 表示字符串及其可选的过期信息
type StringEntry struct {
	Value     string
	ExpiresAt time.Time
	HasExpiry bool
}

// StringStore 实现字符串操作
type StringStore struct {
	sync.RWMutex
	m map[string]StringEntry
}

func NewStringStore() *StringStore {
	return &StringStore{
		m: make(map[string]StringEntry),
	}
}

// checkString 检查键是否存在、类型是否正确以及是否过期
// 必须在调用者持有读锁或写锁的情况下调用
func (s *StringStore) checkString(key string) (StringEntry, bool) {
	entry, exists := s.m[key]
	if !exists {
		return StringEntry{}, false
	}

	if entry.HasExpiry && time.Now().After(entry.ExpiresAt) {
		// 需要写锁来删除过期键，调用者需确保已持有写锁
		delete(s.m, key)
		return StringEntry{}, false
	}

	return entry, true
}

// SetString 设置键值对，并可设置过期时间
func (s *StringStore) SetString(key, value string, expiresAt time.Time, hasExpiry bool) {
	s.Lock()
	defer s.Unlock()

	entry := StringEntry{
		Value:     value,
		ExpiresAt: expiresAt,
		HasExpiry: hasExpiry,
	}
	s.m[key] = entry
}

// GetString 获取字符串值
func (s *StringStore) GetString(key string) (string, bool) {
	s.RLock()
	defer s.RUnlock()

	entry, ok := s.checkString(key)
	if !ok {
		return "", false
	}
	return entry.Value, true
}

func (s *StringStore) Increment(key string) (int, error) {
	s.Lock()
	defer s.Unlock()
	// 检查并处理过期键
	entry, exists := s.checkString(key)

	var value int
	var err error

	if exists {
		value, err = strconv.Atoi(entry.Value)
		if err != nil {
			return 0, errors.New("value is not an integer or out of range")
		}
	} else {
		value = 0
	}
	value++
	newEntry := StringEntry{
		Value:     strconv.Itoa(value),
		ExpiresAt: entry.ExpiresAt,
		HasExpiry: entry.HasExpiry,
	}
	s.m[key] = newEntry
	return value, nil
}
