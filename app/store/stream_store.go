package store

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidIDFormat      = errors.New("Invalid ID format")
	ErrIDTooSmall           = errors.New("The ID specified in XADD must be greater than 0-0")
	ErrIDNotGreaterThanLast = errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
)

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

// parseID 解析ID字符串为毫秒时间和序列号
func parseID(id string) (millis int64, seq int64, err error) {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return 0, 0, ErrInvalidIDFormat
	}
	millis, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, ErrInvalidIDFormat
	}
	seq, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, ErrInvalidIDFormat
	}
	return millis, seq, nil
}

// validateID 验证新ID是否有效
func (s *StreamStore) validateID(key, newID string) error {
	// 解析新ID
	newMillis, newSeq, err := parseID(newID)
	if err != nil {
		return err
	}
	// 检查是否大于0-0
	if newMillis < 0 || (newMillis == 0 && newSeq == 0) {
		return ErrIDTooSmall
	}
	// 获取流的最后一个条目
	entries, exists := s.streams[key]
	if !exists || len(entries) == 0 {
		return nil // 空流，任何大于0-0的ID都有效
	}
	lastEntry := entries[len(entries)-1]
	lastMillis, lastSeq, err := parseID(lastEntry.ID)
	if err != nil {
		return fmt.Errorf("invalid last entry ID: %s", lastEntry.ID)
	}
	// 验证新ID大于最后一个ID
	if newMillis < lastMillis || (newMillis == lastMillis && newSeq <= lastSeq) {
		return ErrIDNotGreaterThanLast
	}
	return nil
}

// generateNextSequence 生成下一个序列号
func (s *StreamStore) generateNextSequence(key string, newMillis int64) (int64, error) {
	entries, exists := s.streams[key]
	if !exists || len(entries) == 0 {
		// 流为空，特殊处理 0 时间部分
		if newMillis == 0 {
			return 1, nil // 0-0 不允许，所以从 1 开始
		}
		return 0, nil
	}
	// 获取最后一个条目
	lastEntry := entries[len(entries)-1]
	lastMillis, lastSeq, err := parseID(lastEntry.ID)
	if err != nil {
		return 0, fmt.Errorf("Invalid last entry ID: %s", lastEntry.ID)
	}
	switch {
	case newMillis > lastMillis:
		return 0, nil // 新时间部分，序列号从 0 开始
	case newMillis == lastMillis:
		return lastSeq + 1, nil // 相同时间部分，序列号递增
	default:
		return 0, ErrIDNotGreaterThanLast // 新时间部分小于最后条目
	}
}

// AddEntry 向指定流添加条目（不存在则创建）
func (s *StreamStore) AddEntry(key, entryID string, fields map[string]string) (string, error) {
	s.Lock()
	defer s.Unlock()

	finalID := entryID

	// 处理完全自动生成序列号的情况 (*)
	if entryID == "*" {
		millisNow := time.Now().UnixMilli()
		// 生成序列号
		seq, err := s.generateNextSequence(key, millisNow)
		if err != nil {
			return "", nil
		}
		finalID = fmt.Sprintf("%d-%d", millisNow, seq)
	} else if parts := strings.Split(entryID, "-"); len(parts) == 2 && parts[1] == "*" {
		// 处理部分自动生成序列号 (millis-*)
		millisPart, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return "", ErrInvalidIDFormat
		}
		// 生成序列号
		seq, err := s.generateNextSequence(key, millisPart)
		if err != nil {
			return "", err
		}
		finalID = fmt.Sprintf("%d-%d", millisPart, seq)
	}

	// 验证ID
	if err := s.validateID(key, finalID); err != nil {
		return "", err
	}

	// 如果流不存在，创建新流
	if _, exists := s.streams[key]; !exists {
		s.streams[key] = make([]StreamEntry, 0)
	}
	entry := StreamEntry{
		ID:     finalID,
		Fields: fields,
	}
	s.streams[key] = append(s.streams[key], entry)
	return finalID, nil
}
