package store

import "fmt"

// ListStore 实现列表操作
type ListStore struct {
	store *Store
}

func NewListStore(s *Store) *ListStore {
	return &ListStore{store: s}
}

// AppendList 追加元素到列表或创建新列表
func (s *ListStore) AppendList(key string, elements []string) (int, error) {
	data, exists := s.store.GetData(key)
	if !exists {
		list := ListEntry(elements)
		s.store.SetData(key, list)
		length := len(list)
		fmt.Printf("ListStore: AppendList key=%s, elements=%v, new list=%v, length=%d\n", key, elements, list, length)
		return length, nil
	}

	list, ok := data.(ListEntry)
	if !ok {
		fmt.Printf("ListStore: AppendList key=%s, invalid type (not a list)\n", key)
		return 0, fmt.Errorf("WRONGTYPE key is not a list")
	}
	list = append(list, elements...)
	s.store.SetData(key, list)
	length := len(list)
	fmt.Printf("ListStore: AppendList key=%s, elements=%v, updated list=%v, length=%d\n", key, elements, list, length)
	return length, nil
}

// GetListRange 获取列表指定范围的元素
func (s *ListStore) GetListRange(key string, start, stop int) ([]string, error) {
	data, exists := s.store.GetData(key)
	if !exists {
		fmt.Printf("ListStore: GetListRange key=%s, exists=false\n", key)
		return []string{}, nil
	}
	list, ok := data.(ListEntry)
	if !ok {
		fmt.Printf("ListStore: GetListRange key=%s, invalid type (not a list)\n", key)
		return nil, fmt.Errorf("WRONGTYPE key is not a list")
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
		fmt.Printf("ListStore: GetListRange key=%s, start=%d, stop=%d, empty (start >= length or start > stop)\n", key, start, stop)
		return []string{}, nil
	}
	if stop >= length {
		stop = length - 1
	}
	result := list[start : stop+1]
	fmt.Printf("ListStore: GetListRange key=%s, start=%d, stop=%d, result=%v\n", key, start, stop, result)
	return result, nil
}

// PrependList 预插入元素到列表或创建新列表
func (s *ListStore) PrependList(key string, elements []string) (int, error) {
	if len(elements) == 0 {
		return 0, fmt.Errorf("no elements provided for PrependList")
	}
	data, exists := s.store.GetData(key)
	if !exists {
		// 键不存在，创建新列表（元素顺序反转以模拟预插入）
		list := make(ListEntry, len(elements))
		for i, elem := range elements {
			list[len(elements)-1-i] = elem
		}
		s.store.SetData(key, list)
		length := len(list)
		fmt.Printf("ListStore: PrependList key=%s, elements=%v, new list=%v, length=%d\n", key, elements, list, length)
		return length, nil
	}
	list, ok := data.(ListEntry)
	if !ok {
		fmt.Printf("ListStore: PrependList key=%s, invalid type (not a list)\n", key)
		return 0, fmt.Errorf("WRONGTYPE key is not a list")
	}
	// 预插入元素（反转 elements 后追加到开头）
	newList := make(ListEntry, len(elements))
	for i, elem := range elements {
		newList[len(elements)-i-1] = elem
	}
	list = append(newList, list...)
	s.store.SetData(key, list)
	length := len(list)
	fmt.Printf("ListStore: PrependList key=%s, elements=%v, updated list=%v, length=%d\n", key, elements, list, length)
	return length, nil
}
