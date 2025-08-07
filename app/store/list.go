package store

import "fmt"

// ListStore 实现列表操作
type ListStore struct {
	store *Store
}

func NewListStore(s *Store) *ListStore {
	return &ListStore{store: s}
}

// checkList 检查键是否存在、类型是否正确以及列表是否为空
// 必须在调用者持有读锁或写锁的情况下调用
func (s *ListStore) checkList(key string) (ListEntry, bool, error) {
	data, exists := s.store.m[key]
	if !exists {
		fmt.Printf("ListStore: checkList key=%s, exists=false\n", key)
		return nil, false, nil
	}
	list, ok := data.(ListEntry)
	if !ok {
		fmt.Printf("ListStore: checkList key=%s, invalid type (not a list)\n", key)
		return nil, false, fmt.Errorf("WRONGTYPE key is not a list")
	}
	if len(list) == 0 {
		fmt.Printf("ListStore: checkList key=%s, list empty\n", key)
		return nil, false, nil
	}
	return list, true, nil
}

// AppendList 追加元素到列表或创建新列表
func (s *ListStore) AppendList(key string, elements []string) (int, error) {
	s.store.Lock()
	defer s.store.Unlock()
	list, ok, err := s.checkList(key)
	if err != nil {
		return 0, err
	}
	if !ok {
		list = ListEntry(elements)
		s.store.m[key] = list
		length := len(list)
		fmt.Printf("ListStore: AppendList key=%s, elements=%v, new list=%v, length=%d\n", key, elements, list, length)
		return length, nil
	}
	list = append(list, elements...)
	s.store.m[key] = list
	length := len(list)
	fmt.Printf("ListStore: AppendList key=%s, elements=%v, updated list=%v, length=%d\n", key, elements, list, length)
	return length, nil
}

// GetListRange 获取列表指定范围的元素
func (s *ListStore) GetListRange(key string, start, stop int) ([]string, error) {
	s.store.RLock()
	defer s.store.RUnlock()
	list, ok, err := s.checkList(key)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []string{}, nil
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
	s.store.Lock()
	defer s.store.Unlock()

	list, ok, err := s.checkList(key)
	if err != nil {
		return 0, err
	}
	if !ok {
		list = make(ListEntry, len(elements))
		for i, elem := range elements {
			list[len(elements)-1-i] = elem
		}
		s.store.m[key] = list
		length := len(list)
		fmt.Printf("ListStore: PrependList key=%s, elements=%v, new list=%v, length=%d\n", key, elements, list, length)
		return length, nil
	}
	// 预插入元素（反转 elements 后追加到开头）
	newList := make(ListEntry, len(elements))
	for i, elem := range elements {
		newList[len(elements)-i-1] = elem
	}
	list = append(newList, list...)
	s.store.m[key] = list
	length := len(list)
	fmt.Printf("ListStore: PrependList key=%s, elements=%v, updated list=%v, length=%d\n", key, elements, list, length)
	return length, nil
}

// GetListLength 获取对应列表的长度
func (s *ListStore) GetListLength(key string) (int, error) {
	s.store.RLock()
	defer s.store.RUnlock()
	list, ok, err := s.checkList(key)
	if err != nil || !ok {
		return 0, err
	}
	length := len(list)
	fmt.Printf("ListStore: GetListLength key=%s, length=%d\n", key, length)
	return length, nil
}

// PopLElement 移除并返回列表的第一个元素
func (s *ListStore) PopLElement(key string, count int) ([]string, bool, error) {
	// 第一阶段：使用读锁检查数据
	s.store.RLock()
	list, ok, err := s.checkList(key)
	if err != nil || !ok {
		s.store.RUnlock()
		return []string{}, ok, err
	}
	s.store.RUnlock()

	// 第二阶段：获取写锁并重新验证
	s.store.Lock()
	defer s.store.Unlock()

	list, ok, err = s.checkList(key)
	if err != nil || !ok {
		return []string{}, ok, err
	}

	length := len(list)
	popCount := count
	if count > length {
		popCount = length
	}

	popped := list[:popCount]
	list = list[popCount:]
	if len(list) == 0 {
		delete(s.store.m, key)
		fmt.Printf("ListStore: PopLElement key=%s, count=%d, popped=%v, list empty, deleted\n", key, count, popped)
	} else {
		s.store.m[key] = list
		fmt.Printf("ListStore: PopLElement key=%s, count=%d, popped=%v, updated list=%v\n", key, count, popped, list)
	}
	return popped, true, nil
}
