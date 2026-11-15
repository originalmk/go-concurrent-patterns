package mkcontext

import (
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
	})
}
