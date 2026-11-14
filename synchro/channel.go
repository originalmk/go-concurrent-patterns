package synchro

func firstValidChannel() {
	setup := make(chan struct{})
	var a int
	var b int
	f := func() {
		a = 1
		b = 2
		close(setup)
	}
	g := func() {
		<-setup
		print(a)
		print(b)
	}
	go f()
	g()
}

func secondValidChannel() {
	var a string
	done := make(chan struct{})
	setup := make(chan struct{})
	go func() {
		<-setup
		a = "hello"
		close(done)
	}()
	doprint := func() {
		select {
		case <-done:
		case setup <- struct{}{}:
			<-done
		}
		print(a)
	}
	go doprint()
	go doprint()
}

func thirdValidChannel() {
	var a string
	done := make(chan struct{})
	setup := func() {
		a = "hello"
		close(done)
	}
	go setup()
	<-done
	print(a)
}
