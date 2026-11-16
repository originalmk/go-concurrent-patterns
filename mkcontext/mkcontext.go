package mkcontext

import (
	"errors"
	"sync"
	"time"

	"krzyzanowski.dev/goconctrain/unboundedchan"
)

var DeadlineExceeded = errors.New("deadline exceeded")
var Canceled = errors.New("canceled")

type afterDef struct {
	mu          sync.Mutex
	f           func()
	aborted     bool
	startedExec bool
}

type MkContext struct {
	done       chan struct{}
	err        error
	errchan    chan error
	erronce    sync.Once
	causeErr   error
	aftersSend chan<- *afterDef
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
	ctx, cancel := makeContext(nil, nil)
	cancel(nil)
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
	aftersRecv, aftersSend, aftersStop := unboundedchan.NewUnboundedChan[*afterDef]()
	mkctx.aftersSend = aftersSend
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
		defer aftersStop(true)
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
		// We stop receiving new after funcs. The user needs to select together with
		// mkctx.done to handle this properly.
		aftersStop(false)
		for afterDef := range aftersRecv {
			afterDef.mu.Lock()
			if afterDef.aborted {
				afterDef.mu.Unlock()
				continue
			}
			afterDef.startedExec = true
			afterDef.mu.Unlock()
			go afterDef.f()
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

func AfterFunc(ctx *MkContext, f func()) (stop func() bool) {
	// This is case for Background(). We just never execute the AfterFunc then.
	if ctx.aftersSend == nil {
		return func() bool { return true }
	}
	def := &afterDef{f: f}
	select {
	case ctx.aftersSend <- def:
		stop = func() bool {
			def.mu.Lock()
			defer def.mu.Unlock()
			if def.startedExec {
				return false
			} else {
				def.aborted = true
				return true
			}
		}
	case <-ctx.done:
		go f()
		stop = func() bool { return false }
	}
	return
}
