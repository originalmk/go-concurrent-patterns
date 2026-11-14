package synchro

import "sync/atomic"

// The semantics here is different, but I don't think atomics
// can handle more complex structures. This is just not for it.
// It removes the race though.
func firstValidAtomics() {
	var a atomic.Int32
	var b atomic.Int32
	f := func() {
		a.Store(1)
		a.Store(2)
	}
	g := func() {
		print(a.Load())
		print(b.Load())
	}
	go f()
	g()
}
