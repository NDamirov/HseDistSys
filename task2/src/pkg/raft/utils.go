package raft

import (
	"fmt"
	"sync"
)

type AtomicInt struct {
	sync.Mutex
	value int
}

func (ai *AtomicInt) Set(value int) {
	ai.Lock()
	defer ai.Unlock()
	ai.value = value
}

func (ai *AtomicInt) Get() int {
	ai.Lock()
	defer ai.Unlock()
	return ai.value
}

func (ai *AtomicInt) Inc(n int) {
	ai.Lock()
	defer ai.Unlock()
	ai.value += n
}

func GetAddress(port int) string {
	return fmt.Sprintf("http://raft%d:%d", port-8080, port)
}
