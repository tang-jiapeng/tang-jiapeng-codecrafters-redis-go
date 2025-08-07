package store

import (
	"fmt"
	"time"
)

// StringStore 实现字符串操作
type StringStore struct {
	store *Store
}

func NewStringStore(s *Store) *StringStore {
	return &StringStore{store: s}
}

// checkString 检查键是否存在、类型是否正确以及是否过期
// 必须在调用者持有读锁或写锁的情况下调用
func (s *StringStore) checkString(key string) (StringEntry, bool) {
	data, exists := s.store.m[key]
	if !exists {
		fmt.Printf("StringStore: checkString key=%s, exists=false\n", key)
		return StringEntry{}, false
	}

	stringEntry, ok := data.(StringEntry)
	if !ok {
		fmt.Printf("StringStore: checkString key=%s, invalid type (not a string)\n", key)
		return StringEntry{}, false
	}

	if stringEntry.HasExpiry && time.Now().After(stringEntry.ExpiresAt) {
		// 需要写锁来删除过期键，调用者需确保已持有写锁
		delete(s.store.m, key)
		fmt.Printf("StringStore: checkString key=%s, expired, deleted\n", key)
		return StringEntry{}, false
	}

	return stringEntry, true
}

// SetString 设置字符串键值对
func (s *StringStore) SetString(key, value string, expiresAt time.Time, hasExpiry bool) {
	s.store.Lock()
	defer s.store.Unlock()

	stringEntry := StringEntry{
		Value:     value,
		ExpiresAt: expiresAt,
		HasExpiry: hasExpiry,
	}
	s.store.m[key] = stringEntry
	fmt.Printf("StringStore: SetString key=%s, value=%s, hasExpiry=%v, expiresAt=%v\n", key, value, hasExpiry, expiresAt)
}

// GetString 获取字符串值
func (s *StringStore) GetString(key string) (string, bool) {
	s.store.RLock()
	defer s.store.RUnlock()

	stringEntry, ok := s.checkString(key)
	if !ok {
		return "", false
	}
	fmt.Printf("StringStore: GetString key=%s, value=%s\n", key, stringEntry.Value)
	return stringEntry.Value, true
}
