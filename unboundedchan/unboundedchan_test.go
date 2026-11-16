package unboundedchan

import "testing"

func TestBasic(t *testing.T) {
	r, w, stop := NewUnboundedChan[int]()
	defer stop(true)
	for i := range 10 {
		w <- i
	}
	close(w)
	for i := range r {
		t.Log(i)
	}
}

func TestDrain(t *testing.T) {
	r, w, stop := NewUnboundedChan[int]()
	defer stop(true)

	for i := range 10 {
		w <- i
	}
	stop(false)
	for i := range r {
		t.Log(i)
	}
}
