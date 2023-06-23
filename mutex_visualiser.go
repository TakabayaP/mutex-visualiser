package mutexvisualiser

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/goccy/go-graphviz"
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
	return []string{"lock_start", "lock", "unlock", "set", "read"}[a]
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

func (m *MutexVisualiser[T]) RenderGraph(path string) {
	sort.Slice(m.actions, func(i, j int) bool {
		return m.actions[i].actionTime.Before(m.actions[j].actionTime)
	})

	g := graphviz.New()
	var b []byte
	var graphStr string

	t := template.New("dot")
	t.Parse(tmplGraph)

	var tmp bytes.Buffer
	var GNos map[uint64]int = make(map[uint64]int)
	GNos[1] = 1

	tNo := 1
	mNo := 1
	for _, v := range m.actions {
		if _, ok := GNos[v.gID]; !ok {
			GNos[v.gID] = 1
			graphStr += createGBranch(v.gID, v.parentGID)
		}
		switch v.actionType {
		case lockStart:
			graphStr += actionOnG("waiting for lock...", v.gID, GNos[v.gID], tNo)
		case read:
			graphStr += actionToM("read", "back", v.gID, GNos[v.gID], tNo, mNo)
			mNo++
		case unlock:
			graphStr += actionToM("unlock", "back", v.gID, GNos[v.gID], tNo, mNo)
			mNo++
		default:
			graphStr += actionToM(v.actionType.String(), "forward", v.gID, GNos[v.gID], tNo, mNo)
			mNo++
		}
		tNo++
		GNos[v.gID]++
	}
	if err := t.Execute(&tmp, graphStr); err != nil {
		log.Fatal(err)
	}
	fmt.Println(tmp.String())
	b = append(b, tmp.Bytes()...)

	graph, err := graphviz.ParseBytes(b)
	if err != nil {
		fmt.Println("Error parsing graphviz bytes", err)
	}
	if err := g.RenderFilename(graph, graphviz.SVG, path); err != nil {
		log.Fatal(err)
	}
}

func createGBranch(GID uint64, PGID uint64) string {
	t := template.New("branch")
	t.Parse(tmplCreateGBranch)
	var tmp bytes.Buffer
	if err := t.Execute(&tmp, map[string]uint64{
		"GID":  GID,
		"PGID": PGID,
	}); err != nil {
		log.Fatal(err)
	}
	return tmp.String()
}

func actionToM(ActionType string, ActionDir string, GID uint64, GNo int, TNo int, MNo int) string {
	t := template.New("set")
	t.Parse(tmplActionToMutex)
	var tmp bytes.Buffer
	if err := t.Execute(&tmp, map[string]string{
		"ActionType": ActionType,
		"ActionDir":  ActionDir,
		"GID":        fmt.Sprint(GID),
		"GNo":        strconv.Itoa(GNo),
		"NextGNo":    strconv.Itoa(GNo + 1),
		"TNo":        strconv.Itoa(TNo),
		"NextTNo":    strconv.Itoa(TNo + 1),
		"MNo":        strconv.Itoa(MNo),
		"NextMNo":    strconv.Itoa(MNo + 1),
	}); err != nil {
		log.Fatal(err)
	}
	return tmp.String()
}

func actionOnG(ActionType string, GID uint64, GNo int, TNo int) string {
	t := template.New("on")
	t.Parse(tmplActionOnG)
	var tmp bytes.Buffer
	if err := t.Execute(&tmp, map[string]string{
		"ActionType": ActionType,
		"GID":        fmt.Sprint(GID),
		"GNo":        strconv.Itoa(GNo),
		"NextGNo":    strconv.Itoa(GNo + 1),
		"TNo":        strconv.Itoa(TNo),
		"NextTNo":    strconv.Itoa(TNo + 1),
	}); err != nil {
		log.Fatal(err)
	}
	return tmp.String()
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
