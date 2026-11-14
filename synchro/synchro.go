package synchro

import "sync"

func firstInvalid() {
	var a int
	var b int
	f := func() {
		a = 1
		b = 2
	}
	g := func() {
		print(a)
		print(b)
	}
	go f()
	g()
}

func secondInvalid() {
	var a string
	var done bool
	var once sync.Once
	setup := func() {
		a = "hello"
		done = true
	}
	doprint := func() {
		if !done {
			once.Do(setup)
		}
		print(a)
	}
	go doprint()
	go doprint()
}

func thirdInvalid() {
	var a string
	var done bool
	setup := func() {
		a = "hello"
		done = true
	}
	go setup()
	for !done {
	}
	print(a)
}
