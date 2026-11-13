package main

import (
	"context"
	"fmt"
	"slices"
	"sync"
)

func NewUnboundedChan[T any]() (recv <-chan T, send chan<- T, stop func()) {
	buf := make([]T, 0)
	r := make(chan T)
	send = r
	s := make(chan T)
	recv = s
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	stop = func() {
		cancel()
		wg.Wait()
	}
	needsNewCandidate := true
	wg.Go(func() {
		niled := 0
		curRecv := r
		var curSend chan T
		var toSend T
		for {
			if needsNewCandidate && len(buf) > 0 {
				toSend = buf[0]
				var t T
				buf[0] = t
				buf = buf[1:]
				niled += 1
				if niled > len(buf) {
					buf = slices.Clone(buf)
					niled = 0
				}
				curSend = s
				needsNewCandidate = false
			}
			if curSend == nil && curRecv == nil {
				close(s)
				return
			}
			select {
			case v, ok := <-curRecv:
				if !ok {
					curRecv = nil
					break
				}
				buf = append(buf, v)
			case curSend <- toSend:
				needsNewCandidate = true
				curSend = nil
			case <-ctx.Done():
				return
			}
		}
	})
	return
}

func main() {
	r, w, stop := NewUnboundedChan[int]()
	defer stop()
	for i := range 10 {
		w <- i
	}
	close(w)
	for i := range r {
		fmt.Println(i)
	}
}
