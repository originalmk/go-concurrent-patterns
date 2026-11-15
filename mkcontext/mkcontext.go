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
	ctx, _ := makeContext(nil, nil)
	return ctx
}

func WithCancel(parent MkContext) (mkctx MkContext, cancel CancelFunc) {
	return makeContext(&parent, nil)
}

func WithDeadline(parent MkContext, deadline time.Time) (mkctx MkContext, cancel CancelFunc) {
	return makeContext(&parent, &deadline)
}

func makeContext(parent *MkContext, deadline *time.Time) (mkctx MkContext, cancel CancelFunc) {
	if parent == nil && deadline == nil {
		return MkContext{
			done: make(chan struct{}),
		}, func() {}
	}

	mkctx = Background()
	c := make(chan struct{})
	var o sync.Once
	cancel = func() {
		o.Do(func() {
			close(c)
			<-mkctx.done
		})
	}
	var pch chan struct{}
	if parent != nil {
		pch = parent.done
	}
	var dch <-chan time.Time
	if deadline != nil {
		dch = time.After(time.Until(*deadline))
	}
	go func() {
		select {
		case <-c:
		case <-pch:
		case <-dch:
		}
		close(mkctx.done)
	}()
	return
}
