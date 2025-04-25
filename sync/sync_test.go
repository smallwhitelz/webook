package sync

import (
	"sync"
	"testing"
)

func TestSyncMap(t *testing.T) {
	m := &sync.Map{}
	// 写入
	m.Store("key1", "value1")
	// 读取，注意 Load 的第一个返回值是 any 类型
	val, ok := m.Load("key1")
	if ok {
		t.Log(val.(string))
	}

	// 不存在就写入，存在就返回原始的值
	val, loaded := m.LoadOrStore("key1", "value2")
	if loaded {
		t.Log("加载到数据", val.(string))
	}

	val, loaded = m.LoadOrStore("key2", "value12")
	if !loaded {
		t.Log("没加载到数据，使用新数据", val.(string))
	}

	// 线程安全的 CAS 操作
	swapped := m.CompareAndSwap("key1", "value1", "value3")
	if swapped {
		t.Log("将原本的 value1 替换为 value3")
	}
}
