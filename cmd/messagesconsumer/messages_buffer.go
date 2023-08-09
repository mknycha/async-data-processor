package messagesconsumer

import "sync"

type messagesBuffer struct {
	rmu      sync.RWMutex
	messages [][]byte
}

func (csc *messagesBuffer) Push(message []byte) {
	csc.rmu.Lock()
	csc.messages = append(csc.messages, message)
	csc.rmu.Unlock()
}

func (csc *messagesBuffer) GetMessages() [][]byte {
	csc.rmu.RLock()
	defer csc.rmu.RUnlock()
	return csc.messages
}

func (csc *messagesBuffer) Reset() {
	csc.rmu.Lock()
	defer csc.rmu.Unlock()
	csc.messages = make([][]byte, 0)
}

func (csc *messagesBuffer) Size() int {
	csc.rmu.RLock()
	defer csc.rmu.RUnlock()
	return len(csc.messages)
}
