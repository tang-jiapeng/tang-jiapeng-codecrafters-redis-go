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

// SetString 设置字符串键值对
func (s *StringStore) SetString(key, value string, expiresAt time.Time, hasExpiry bool) {
	s.store.SetData(key, StringEntry{
		Value:     value,
		ExpiresAt: expiresAt,
		HasExpiry: hasExpiry,
	})
	fmt.Printf("StringStore: SetString key=%s, value=%s, hasExpiry=%v, expiresAt=%v\n", key, value, hasExpiry, expiresAt)
}

// GetString 获取字符串值
func (s *StringStore) GetString(key string) (string, bool) {
	data, exists := s.store.GetData(key)
	if !exists {
		fmt.Printf("StringStore: GetString key=%s, exists=false\n", key)
		return "", false
	}

	stringEntry, ok := data.(StringEntry)
	if !ok {
		fmt.Printf("StringStore: GetString key=%s, invalid type (not a string)\n", key)
		return "", false
	}

	if stringEntry.HasExpiry && time.Now().After(stringEntry.ExpiresAt) {
		s.store.Delete(key)
		fmt.Printf("StringStore: GetString key=%s, expired, deleted\n", key)
		return "", false
	}
	fmt.Printf("StringStore: GetString key=%s, value=%s\n", key, stringEntry.Value)
	return stringEntry.Value, true
}
