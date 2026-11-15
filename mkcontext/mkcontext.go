package mkcontext

import (
	"errors"
	"sync"
	"time"
)

var DeadlineExceeded = errors.New("deadline exceeded")
var Canceled = errors.New("canceled")

type MkContext struct {
	done     chan struct{}
	err      error
	errchan  chan error
	erronce  sync.Once
	causeErr error
}

type CancelFunc func()
type CancelCauseFunc func(error)

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

func (mkctx *MkContext) cause() error {
	select {
	case <-mkctx.done:
		if mkctx.causeErr != nil {
			return mkctx.causeErr
		} else {
			return mkctx.Err()
		}
	default:
		return nil
	}
}

func Background() *MkContext {
	ctx, _ := makeContext(nil, nil)
	return ctx
}

func WithCancel(parent *MkContext) (mkctx *MkContext, cancel CancelFunc) {
	return convertCancel(makeContext(parent, nil))
}

func WithDeadline(parent *MkContext, deadline time.Time) (mkctx *MkContext, cancel CancelFunc) {
	return convertCancel(makeContext(parent, &deadline))
}

func WithCancelCause(parent *MkContext) (*MkContext, CancelCauseFunc) {
	return makeContext(parent, nil)
}

func convertCancel(mkctx *MkContext, cancelCause CancelCauseFunc) (*MkContext, CancelFunc) {
	return mkctx, func() { cancelCause(Canceled) }
}

func makeContext(parent *MkContext, deadline *time.Time) (mkctx *MkContext, cancel CancelCauseFunc) {
	if parent == nil && deadline == nil {
		return &MkContext{
			done:    make(chan struct{}),
			errchan: make(chan error, 1),
		}, func(error) {}
	}

	mkctx = Background()
	causeChan := make(chan error, 1)
	var o sync.Once
	cancel = func(cause error) {
		o.Do(func() {
			causeChan <- cause
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
		case cause := <-causeChan:
			if cause == nil {
				mkctx.causeErr = Canceled
			} else {
				mkctx.causeErr = cause
			}
			mkctx.errchan <- Canceled
		case <-pch:
			mkctx.causeErr = parent.cause()
			mkctx.errchan <- Canceled
		case <-dch:
			mkctx.causeErr = DeadlineExceeded
			mkctx.errchan <- DeadlineExceeded
		}
	}()
	return
}

func Cause(ctx *MkContext) error {
	select {
	case <-ctx.Done():
		return ctx.causeErr
	default:
		return nil
	}
}
