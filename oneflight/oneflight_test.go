package oneflight

import (
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"golang.org/x/sync/singleflight"
)

type flighter interface {
	Do(key string, f func() (any, error)) (any, error, bool)
}

func TestSingleflight(t *testing.T) {
	testImpl(t, func() flighter { return new(singleflight.Group) })
}

func TestOneflight(t *testing.T) {
	testImpl(t, func() flighter { return new(Group) })
}

func testImpl(t *testing.T, cons func() flighter) {
	t.Run("sequential", func(t *testing.T) {
		testSequential(t, cons())
	})
	t.Run("shared", func(t *testing.T) {
		testShared(t, cons())
	})
}

func testSequential(t *testing.T, fl flighter) {
	v1, err1, shared1 := fl.Do("t1", func() (any, error) {
		return 0, nil
	})
	v2, err2, shared2 := fl.Do("t2", func() (any, error) {
		return 0, nil
	})
	if v1 != 0 || v2 != 0 {
		t.Error()
	}
	if err1 != nil || err2 != nil {
		t.Error()
	}
	if shared1 != false || shared2 != false {
		t.Error()
	}
}

func testShared(t *testing.T, fl flighter) {
	synctest.Test(t, func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Go(func() {
			v1, err1, shared1 := fl.Do("t1", func() (any, error) {
				time.Sleep(time.Second + time.Nanosecond)
				return 0, nil
			})
			if v1 != 0 {
				t.Error()
			}
			if err1 != nil {
				t.Error()
			}
			if !shared1 {
				t.Error()
			}
		})
		wg.Go(func() {
			v2, err2, shared2 := fl.Do("t1", func() (any, error) {
				time.Sleep(time.Second)
				return 0, nil
			})
			if v2 != 0 {
				t.Error()
			}
			if err2 != nil {
				t.Error()
			}
			if !shared2 {
				t.Error()
			}
		})
		wg.Wait()
	})
}
