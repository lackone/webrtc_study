package ws

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// Listener 是事件监听器函数类型
type Listener[T any] func(data T)

// listenerEntry 封装监听器及其元信息
type listenerEntry[T any] struct {
	id   uint64
	fn   Listener[T]
	once bool
}

// Token 用于标识一个监听器，可用于 Off()
type Token uint64

// Emitter 是事件派发器
type Emitter[T any] struct {
	mu        sync.RWMutex
	listeners map[string][]*listenerEntry[T]
	nextID    uint64 // 用于生成唯一监听器 ID
}

// New 创建一个新的类型安全事件派发器
func NewEmitter[T any]() *Emitter[T] {
	return &Emitter[T]{
		listeners: make(map[string][]*listenerEntry[T]),
	}
}

// On 注册一个普通监听器，返回 Token 用于后续移除
func (e *Emitter[T]) On(event string, fn Listener[T]) Token {
	if event == "" {
		panic("event name cannot be empty")
	}
	if reflect.ValueOf(fn).IsNil() {
		panic("listener function cannot be nil")
	}

	id := atomic.AddUint64(&e.nextID, 1)
	entry := &listenerEntry[T]{id: id, fn: fn, once: false}

	e.mu.Lock()
	e.listeners[event] = append(e.listeners[event], entry)
	e.mu.Unlock()

	return Token(id)
}

// Once 注册一个一次性监听器，触发后自动移除
func (e *Emitter[T]) Once(event string, fn Listener[T]) Token {
	if event == "" {
		panic("event name cannot be empty")
	}
	if reflect.ValueOf(fn).IsNil() {
		panic("listener function cannot be nil")
	}

	id := atomic.AddUint64(&e.nextID, 1)
	entry := &listenerEntry[T]{id: id, fn: fn, once: true}

	e.mu.Lock()
	e.listeners[event] = append(e.listeners[event], entry)
	e.mu.Unlock()

	return Token(id)
}

// Off 通过 Token 移除监听器（推荐方式）
func (e *Emitter[T]) Off(event string, token Token) bool {
	if event == "" {
		return false
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	list, ok := e.listeners[event]
	if !ok {
		return false
	}

	for i, entry := range list {
		if entry.id == uint64(token) {
			// 从切片中移除
			e.listeners[event] = append(list[:i], list[i+1:]...)
			return true
		}
	}
	return false
}

// OffFunc 移除指定函数的监听器（仅适用于具名函数或相同变量引用）
// ⚠️ 对匿名函数无效，建议优先使用 Token 方式
func (e *Emitter[T]) OffFunc(event string, fn Listener[T]) bool {
	if event == "" || reflect.ValueOf(fn).IsNil() {
		return false
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	list, ok := e.listeners[event]
	if !ok {
		return false
	}

	for i, entry := range list {
		if reflect.ValueOf(entry.fn).Pointer() == reflect.ValueOf(fn).Pointer() {
			e.listeners[event] = append(list[:i], list[i+1:]...)
			return true
		}
	}
	return false
}

// Emit 触发事件，按注册顺序调用监听器
// 若监听器 panic，会被 recover 并打印错误（可替换为自定义错误处理器）
func (e *Emitter[T]) Emit(event string, data T) {
	e.mu.RLock()
	// 深拷贝监听器列表，避免在遍历时被修改（尤其是 Once 触发后）
	currentList := make([]*listenerEntry[T], len(e.listeners[event]))
	copy(currentList, e.listeners[event])
	e.mu.RUnlock()

	var onceToRemove []*listenerEntry[T]
	var onceLock sync.Mutex
	var wg sync.WaitGroup

	wg.Add(len(currentList))
	for _, entry := range currentList {
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// 可替换为日志库或自定义错误处理
					fmt.Printf("[Emitter] Listener panic on event %q: %v\n", event, r)
					fmt.Printf("[Emitter] Stack trace:\n%s\n", debug.Stack())
				}
			}()

			entry.fn(data)

			if entry.once {
				onceLock.Lock()
				onceToRemove = append(onceToRemove, entry)
				onceLock.Unlock()
			}
		}()
	}
	wg.Wait()

	// 清理 Once 监听器
	if len(onceToRemove) > 0 {
		e.mu.Lock()
		list := e.listeners[event]
		filtered := make([]*listenerEntry[T], 0, len(list))

		for _, entry := range list {
			shouldRemove := false
			for _, toRemove := range onceToRemove {
				if entry.id == toRemove.id {
					shouldRemove = true
					break
				}
			}
			if !shouldRemove {
				filtered = append(filtered, entry)
			}
		}
		e.listeners[event] = filtered
		e.mu.Unlock()
	}
}

// RemoveAllListeners 移除某个事件的所有监听器
func (e *Emitter[T]) RemoveAllListeners(event string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.listeners, event)
}

// EventNames 返回当前已注册的所有事件名称
func (e *Emitter[T]) EventNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	names := make([]string, 0, len(e.listeners))
	for name := range e.listeners {
		names = append(names, name)
	}
	return names
}

// ListenerCount 返回某事件的监听器数量
func (e *Emitter[T]) ListenerCount(event string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.listeners[event])
}
