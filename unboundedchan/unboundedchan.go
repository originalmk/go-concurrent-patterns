package unboundedchan

import (
	"context"
	"slices"
	"sync"
)

// Call stop(true) to hard-stop or stop(false) to be able to get remaining messages.
// If you call stop(false), you still need to call stop(true).
func NewUnboundedChan[T any]() (recv <-chan T, send chan<- T, stop func(bool)) {
	buf := make([]T, 0)
	r := make(chan T)
	send = r
	s := make(chan T)
	recv = s
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	softclose := make(chan struct{})
	stop = func(hard bool) {
		if hard {
			cancel()
			wg.Wait()
		} else {
			close(softclose)
		}
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
			case <-softclose:
				curRecv = nil
			case curSend <- toSend:
				needsNewCandidate = true
				curSend = nil
			case <-ctx.Done():
				close(s)
				return
			}
		}
	})
	return
}
