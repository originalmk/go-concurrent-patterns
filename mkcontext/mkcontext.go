package mkcontext

import (
	"errors"
	"sync"
	"time"
)

var DeadlineExceeded = errors.New("deadline exceeded")
var Canceled = errors.New("canceled")

type MkContext struct {
	done    chan struct{}
	err     error
	errchan chan error
	erronce sync.Once
}

type CancelFunc func()

func (mkctx *MkContext) Done() <-chan struct{} {
	return mkctx.done
}

func (mkctx *MkContext) Err() error {
	select {
	case <-mkctx.done:
		mkctx.erronce.Do(func() {
			reason := <-mkctx.errchan
			mkctx.err = reason
		})
	default:
	}
	return mkctx.err
}

func Background() *MkContext {
	ctx, _ := makeContext(nil, nil)
	return ctx
}

func WithCancel(parent *MkContext) (mkctx *MkContext, cancel CancelFunc) {
	return makeContext(parent, nil)
}

func WithDeadline(parent *MkContext, deadline time.Time) (mkctx *MkContext, cancel CancelFunc) {
	return makeContext(parent, &deadline)
}

func makeContext(parent *MkContext, deadline *time.Time) (mkctx *MkContext, cancel CancelFunc) {
	if parent == nil && deadline == nil {
		return &MkContext{
			done:    make(chan struct{}),
			errchan: make(chan error, 1),
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
		defer close(mkctx.done)
		select {
		case <-c:
			mkctx.errchan <- Canceled
		case <-pch:
			mkctx.errchan <- Canceled
		case <-dch:
			mkctx.errchan <- DeadlineExceeded
		}
	}()
	return
}
