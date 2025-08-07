package store

import (
	"fmt"
	"sync"
)

// ListOps 定义列表操作接口
type ListOps interface {
	AppendList(key string, elements []string) (int, error)
	PrependList(key string, elements []string) (int, error)
	GetListRange(key string, start, stop int) ([]string, error)
	GetListLength(key string) (int, error)
	PopLElement(key string, count int) ([]string, bool, error)
}

// ListStore 实现列表操作
type ListStore struct {
	sync.RWMutex
	m map[string][]string
}

func NewListStore() *ListStore {
	return &ListStore{
		m: make(map[string][]string),
	}
}

// checkList 检查键是否存在、类型是否正确以及列表是否为空
// 必须在调用者持有读锁或写锁的情况下调用
func (s *ListStore) checkList(key string) ([]string, bool, error) {
	list, exists := s.m[key]
	if !exists || len(list) == 0 {
		return nil, false, nil
	}
	return list, true, nil
}

// AppendList 追加元素到列表或创建新列表
func (s *ListStore) AppendList(key string, elements []string) (int, error) {
	s.Lock()
	defer s.Unlock()
	list, ok, err := s.checkList(key)
	if err != nil {
		return 0, err
	}
	if !ok {
		s.m[key] = elements
	} else {
		list = append(list, elements...)
		s.m[key] = list
	}
	return len(s.m[key]), nil
}

// GetListRange 获取列表指定范围的元素
func (s *ListStore) GetListRange(key string, start, stop int) ([]string, error) {
	s.RLock()
	defer s.RUnlock()
	list, ok, err := s.checkList(key)
	if err != nil || !ok {
		return []string{}, err
	}
	length := len(list)
	// 处理负索引
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0 // 负索引超出范围，设置为 0
		}
	}
	if stop < 0 {
		stop = length + stop
		if stop < 0 {
			stop = 0 // 负索引超出范围，设置为 0
		}
	}

	// 处理范围
	if start >= length || start > stop {
		return []string{}, nil
	}
	if stop >= length {
		stop = length - 1
	}
	result := list[start : stop+1]
	return result, nil
}

// PrependList 预插入元素到列表或创建新列表
func (s *ListStore) PrependList(key string, elements []string) (int, error) {
	if len(elements) == 0 {
		return 0, fmt.Errorf("no elements provided for PrependList")
	}
	s.Lock()
	defer s.Unlock()

	list, ok, err := s.checkList(key)
	if err != nil {
		return 0, err
	}
	newList := make([]string, len(elements))
	for i, val := range elements {
		newList[len(elements)-1-i] = val
	}
	if !ok {
		s.m[key] = newList
	} else {
		s.m[key] = append(newList, list...)
	}
	return len(s.m[key]), nil
}

// GetListLength 获取对应列表的长度
func (s *ListStore) GetListLength(key string) (int, error) {
	s.RLock()
	defer s.RUnlock()
	list, ok, err := s.checkList(key)
	if err != nil || !ok {
		return 0, err
	}
	length := len(list)
	return length, nil
}

// PopLElement 移除并返回列表的第一个元素
func (s *ListStore) PopLElement(key string, count int) ([]string, bool, error) {
	// 第一阶段：使用读锁检查数据
	s.RLock()
	list, ok, err := s.checkList(key)
	s.RUnlock()
	if err != nil || !ok {
		return []string{}, ok, err
	}

	// 第二阶段：获取写锁并重新验证
	s.Lock()
	defer s.Unlock()
	list, ok, err = s.checkList(key)
	if err != nil || !ok {
		return []string{}, ok, err
	}
	if count > len(list) {
		count = len(list)
	}
	popped := list[:count]
	remaining := list[count:]
	if len(remaining) == 0 {
		delete(s.m, key)
	} else {
		s.m[key] = remaining
	}
	return popped, true, nil
}
