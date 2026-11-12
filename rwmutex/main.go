package main

import "sync"

type RWMutex struct {
	smu              sync.Mutex
	rCounter         int
	wCounter         int
	readersDecreased chan struct{}
	writersDecreased chan struct{}
	writersWaiting   chan struct{}
	wWaiting         int
}

func NewRWMutex() *RWMutex {
	return &RWMutex{
		readersDecreased: make(chan struct{}),
		writersDecreased: make(chan struct{}),
		writersWaiting:   make(chan struct{}),
	}
}

func (rwm *RWMutex) Lock() {
	rwm.smu.Lock()

	for rwm.wWaiting != 0 {
		ch := rwm.writersWaiting
		rwm.smu.Unlock()
		<-ch
		rwm.smu.Lock()
	}

	rwm.wWaiting += 1

	for rwm.rCounter != 0 {
		ch := rwm.readersDecreased
		rwm.smu.Unlock()
		<-ch
		rwm.smu.Lock()
	}

	for rwm.wCounter != 0 {
		ch := rwm.writersDecreased
		rwm.smu.Unlock()
		<-ch
		rwm.smu.Lock()
	}

	rwm.wCounter += 1
	rwm.wWaiting -= 1
	close(rwm.writersWaiting)
	rwm.writersWaiting = make(chan struct{})

	rwm.smu.Unlock()
}

func (rwm *RWMutex) RLock() {
	rwm.smu.Lock()

	for rwm.wWaiting != 0 || rwm.wCounter != 0 {
		ch := rwm.writersDecreased
		rwm.smu.Unlock()
		<-ch
		rwm.smu.Lock()
	}

	rwm.rCounter += 1
	rwm.smu.Unlock()
}

func (rwm *RWMutex) Unlock() {
	rwm.smu.Lock()
	rwm.wCounter -= 1
	close(rwm.writersDecreased)
	rwm.writersDecreased = make(chan struct{})
	rwm.smu.Unlock()
}

func (rwm *RWMutex) RUnlock() {
	rwm.smu.Lock()
	rwm.rCounter -= 1
	close(rwm.readersDecreased)
	rwm.readersDecreased = make(chan struct{})
	rwm.smu.Unlock()
}

func main() {
	mu := NewRWMutex()

	mu.Lock()
	mu.Unlock()

	mu.RLock()
	mu.RLock()
	mu.RLock()
	mu.RLock()
	mu.RUnlock()
	mu.RUnlock()
	mu.RUnlock()
	mu.RUnlock()

	//mu.RLock()
	//mu.Lock()
	//mu.RLock()
}
