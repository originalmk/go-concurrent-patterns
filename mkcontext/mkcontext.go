package mkcontext

import "time"
import "sync"

type MkContext struct {
	done chan struct{}
}

type CancelFunc func()

func (mkctx MkContext) Done() <-chan struct{} {
	return mkctx.done
}

func Background() MkContext {
	return MkContext{
		done: make(chan struct{}),
	}
}

func WithCancel(parent MkContext) (mkctx MkContext, cancel CancelFunc) {
	mkctx = Background()
	c := make(chan struct{})
	var o sync.Once
	cancel = func() {
		o.Do(func() {
			close(c)
			<-mkctx.done
		})
	}
	go func() {
		select {
		case <-c:
		case <-parent.done:
		}
		close(mkctx.done)
	}()
	return
}

func WithDeadline(parent MkContext, deadline time.Time) (mkctx MkContext, cancel CancelFunc) {
	mkctx, cancel = WithCancel(parent)
	go func() {
		select {
		case <-mkctx.done:
		case <-time.After(time.Until(deadline)):
			cancel()
		}
	}()
	return
}
