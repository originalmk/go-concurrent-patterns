package synchro

import "sync"

func firstValidMutex() {
	var mu sync.Mutex
	var a int
	var b int
	f := func() {
		mu.Lock()
		a = 1
		b = 2
		mu.Unlock()
	}
	g := func() {
		mu.Lock()
		print(a)
		print(b)
		mu.Unlock()
	}
	go f()
	g()
}

// No mutex needed here, because we have once.
func secondValidMutex() {
	var a string
	var once sync.Once
	setup := func() {
		a = "hello"
	}
	doprint := func() {
		once.Do(setup)
		print(a)
	}
	go doprint()
	go doprint()
}

func thirdValidMutex() {
	var mu sync.Mutex
	var a string
	var done bool
	setup := func() {
		mu.Lock()
		defer mu.Unlock()
		a = "hello"
		done = true
	}
	go setup()
	mu.Lock()
	for !done {
		mu.Unlock()
		mu.Lock()
	}
	print(a)
	mu.Unlock()
}
