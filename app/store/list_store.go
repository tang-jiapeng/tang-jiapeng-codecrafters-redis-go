package store

import (
	"fmt"
	"sync"
	"time"
)

// ListOps 定义列表操作接口
type ListOps interface {
	AppendList(key string, elements []string) (int, error)
	PrependList(key string, elements []string) (int, error)
	GetListRange(key string, start, stop int) ([]string, error)
	GetListLength(key string) (int, error)
	LPopElement(key string, count int) ([]string, bool, error)
	BLPopElement(key string, timeout time.Duration) (string, bool, error)
}

// ListStore 实现列表操作
type ListStore struct {
	sync.RWMutex
	m       map[string][]string
	waiters map[string][]chan struct{} // 每个键对应的等待队列（通道切片）
}

func NewListStore() *ListStore {
	return &ListStore{
		m:       make(map[string][]string),
		waiters: make(map[string][]chan struct{}),
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
	// 唤醒该键的第一个等待者
	if waiters, ok := s.waiters[key]; ok && len(waiters) > 0 {
		waiter := waiters[0]         // 获取最先等待的客户端
		s.waiters[key] = waiters[1:] // 移除已唤醒的客户端
		if len(s.waiters[key]) == 0 {
			delete(s.waiters, key)
		}
		close(waiter) // 唤醒客户端（非阻塞）
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
	// 唤醒该键的第一个等待者
	if waiters, ok := s.waiters[key]; ok && len(waiters) > 0 {
		waiter := waiters[0]         // 获取最先等待的客户端
		s.waiters[key] = waiters[1:] // 移除已唤醒的客户端
		if len(s.waiters[key]) == 0 {
			delete(s.waiters, key)
		}
		close(waiter) // 唤醒客户端（非阻塞）
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

// LPopElement 移除并返回列表的第一个元素
func (s *ListStore) LPopElement(key string, count int) ([]string, bool, error) {
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

// BLPopElement 阻塞弹出列表头部元素，timeout 0 表示无限阻塞
func (s *ListStore) BLPopElement(key string, timeout time.Duration) (string, bool, error) {
	s.Lock()
	// 快速检查：如果列表有元素，直接弹出
	if list, ok := s.m[key]; ok && len(list) > 0 {
		elem := list[0]
		if len(list) == 1 {
			delete(s.m, key)
		} else {
			s.m[key] = list[1:]
		}
		s.Unlock()
		return elem, true, nil
	}

	// 创建等待通道并加入队列
	ch := make(chan struct{})
	s.waiters[key] = append(s.waiters[key], ch)
	s.Unlock()

	// 处理超时
	var timeoutCh <-chan time.Time
	if timeout > 0 {
		timeoutCh = time.After(timeout)
	}

	select {
	case <-ch: // 被唤醒
		s.Lock()
		defer s.Unlock()
		// 再次检查列表状态
		list, ok := s.m[key]
		if !ok || len(list) == 0 {
			return "", false, nil // 元素已被其他消费者取走
		}
		elem := list[0]
		if len(list) == 1 {
			delete(s.m, key)
		} else {
			s.m[key] = list[1:]
		}
		return elem, true, nil
	case <-timeoutCh: // 超时
		s.Lock()
		defer s.Unlock()
		// 从等待队列中移除自己
		if waiters, ok := s.waiters[key]; ok {
			for i, waiter := range waiters {
				if waiter == ch {
					s.waiters[key] = append(waiters[:i], waiters[i+1:]...)
					if len(s.waiters[key]) == 0 {
						delete(s.waiters, key)
					}
					break
				}
			}
		}
		return "", false, nil
	}
}
