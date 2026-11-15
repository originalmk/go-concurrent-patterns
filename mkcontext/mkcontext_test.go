package mkcontext

import (
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

func TestCancel(t *testing.T) {
	ctx := Background()
	ctxa, cancela := WithCancel(ctx)
	ctxb, _ := WithCancel(ctxa)
	cancela()
	<-ctxa.Done()
	<-ctxb.Done()
}

func TestDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := WithDeadline(Background(), time.Now().Add(time.Second*5))
		defer cancel()
		select {
		case <-ctx.Done():
		case <-time.After(time.Second*5 + time.Nanosecond):
			t.Fatal("context did not get canceled after deadline")
		}
		err := ctx.Err()
		if err != DeadlineExceeded {
			t.Fatal("expected to get deadline exceeded error")
		}
	})
}

func TestErrBackground(t *testing.T) {
	ctx := Background()
	err := ctx.Err()
	if err != nil {
		t.Fatal("expected error to be nil")
	}
}

func TestErrCancel(t *testing.T) {
	ctx, cancel := WithCancel(Background())
	cancel()
	err := ctx.Err()
	if err != Canceled {
		t.Fatal("expected error to be canceled")
	}

}

func TestCause(t *testing.T) {
	helloErr := errors.New("Hello")
	ctxa, cancelCause := WithCancelCause(Background())
	ctxb, cancel := WithCancel(ctxa)
	defer cancel()
	cancelCause(helloErr)
	<-ctxa.Done()
	err := Cause(ctxa)
	if err != helloErr {
		t.Fatal("expected to get hello error as cause")
	}
	<-ctxb.Done()
	err = Cause(ctxb)
	if err != helloErr {
		t.Fatalf("expected to get hello error as cause, got %s", err)
	}
}
