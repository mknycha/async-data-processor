package workerpool

import "sync"

type messagesBatch struct {
	rmu      sync.RWMutex
	messages [][]byte
}

func (csc *messagesBatch) Push(message []byte) {
	csc.rmu.Lock()
	csc.messages = append(csc.messages, message)
	csc.rmu.Unlock()
}

func (csc *messagesBatch) GetMessages() [][]byte {
	csc.rmu.RLock()
	defer csc.rmu.RUnlock()
	return csc.messages
}

func (csc *messagesBatch) Reset() {
	csc.rmu.Lock()
	defer csc.rmu.Unlock()
	csc.messages = make([][]byte, 0)
}

func (csc *messagesBatch) Size() int {
	csc.rmu.RLock()
	defer csc.rmu.RUnlock()
	return len(csc.messages)
}
