package store

import "sync"

// StreamOps 定义stream流操作接口
type StreamOps interface {
	Exists(key string) bool
	AddEntry(key, entryID string, fields map[string]string) (string, error)
}

// StreamEntry 表示流中的一个条目
type StreamEntry struct {
	ID     string
	Fields map[string]string
}

// StreamStore 存储流数据
type StreamStore struct {
	sync.RWMutex
	streams map[string][]StreamEntry
}

func NewStreamStore() *StreamStore {
	return &StreamStore{
		streams: make(map[string][]StreamEntry),
	}
}

// Exists 检查流是否存在
func (s *StreamStore) Exists(key string) bool {
	s.RLock()
	defer s.RUnlock()
	_, exists := s.streams[key]
	return exists
}

// AddEntry 向指定流添加条目（不存在则创建）
func (s *StreamStore) AddEntry(key, entryID string, fields map[string]string) (string, error) {
	s.Lock()
	defer s.Unlock()
	// 如果流不存在，创建新流
	if _, exists := s.streams[key]; exists {
		s.streams[key] = make([]StreamEntry, 0)
	}
	entry := StreamEntry{
		ID:     entryID,
		Fields: fields,
	}
	s.streams[key] = append(s.streams[key], entry)
	return entryID, nil
}
