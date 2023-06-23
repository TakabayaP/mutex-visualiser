package mutexvisualiser

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type Mutex[T any] interface {
	sync.Locker
	Set(T)
	Read() T
}

type MutexVisualiser[T any] struct {
	m       sync.Mutex
	actions []action
	value   T
}

type action struct {
	actionType  actionType
	actionTime  time.Time
	gID         uint64
	parentGID   uint64
	funcName    string
	gParentFunc string
}
type actionType int

const (
	lockStart actionType = iota
	lockEnd
	unlock
	set
	read
)

func (a actionType) String() string {
	return []string{"lock_start", "lock_end", "unlock", "set", "read"}[a]
}

func (m *MutexVisualiser[T]) PrintAll() {
	for _, act := range m.actions {
		fmt.Printf("Action: %s\nAction Time: %s\nGoroutine ID: %d\nParent Goroutine ID: %d\nFunction name: %s\nParent function name: %s\n\n", act.actionType.String(), act.actionTime.Format(timeFormat), act.gID, act.parentGID, act.funcName, act.gParentFunc)
	}
}

func (m *MutexVisualiser[T]) Lock() {
	start := time.Now()
	m.addAction(action{actionType: lockStart,
		actionTime: start})
	m.m.Lock()
	end := time.Now()
	m.addAction(action{actionType: lockEnd,
		actionTime: end,
	})
}

func (m *MutexVisualiser[T]) Unlock() {
	start := time.Now()
	m.m.Unlock()
	m.addAction(action{actionType: unlock,
		actionTime: start,
	})
}

func (m *MutexVisualiser[T]) Read() T {
	m.addAction(action{actionType: read,
		actionTime: time.Now(),
	})
	return m.value
}

func (m *MutexVisualiser[T]) Set(t T) {
	m.addAction(action{actionType: set,
		actionTime: time.Now(),
	})
	m.value = t
}

const timeFormat = "15:04:05.000000"

func (m *MutexVisualiser[T]) addAction(act action) {
	b := make([]byte, 1024)
	b = b[:runtime.Stack(b, false)]
	act.gID = getGID(b)
	act.parentGID = getParentGID(b)
	pc, _, _, _ := runtime.Caller(2)
	act.funcName = runtime.FuncForPC(pc).Name()
	act.gParentFunc = getGParentFunc(b)
	m.actions = append(m.actions, act)
	fmt.Printf("Action: %s\nAction Time: %s\nGoroutine ID: %d\nParent Goroutine ID: %d\nFunction name: %s\nParent function name: %s\n\n", act.actionType.String(), act.actionTime.Format(timeFormat), act.gID, act.parentGID, act.funcName, act.gParentFunc)
}
func getGID(stack []byte) uint64 {
	b := bytes.TrimPrefix(stack, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func getParentGID(stack []byte) uint64 {
	if i := bytes.Index(stack, []byte("created by ")); i >= 0 {
		stack = stack[i:]
	} else {
		return 0
	}
	b := stack[bytes.Index(stack, []byte("in goroutine ")):]
	b = bytes.TrimPrefix(b, []byte("in goroutine "))
	b = b[:bytes.IndexByte(b, '\n')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func getGParentFunc(stack []byte) string {
	if i := bytes.Index(stack, []byte("created by ")); i >= 0 {
		stack = stack[i:]
	} else {
		return "call in main"
	}

	stack = bytes.TrimPrefix(stack, []byte("created by "))
	stack = stack[:bytes.IndexByte(stack, ' ')]
	return string(stack)
}
