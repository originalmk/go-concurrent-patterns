package oneflight

import (
	"sync"
	"testing"
	"testing/synctest"

	"golang.org/x/sync/singleflight"
)

type flighter interface {
	Do(key string, f func() (any, error)) (any, error, bool)
	Forget(key string)
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
		testShared(t, cons(), false)
	})
	t.Run("forget", func(t *testing.T) {
		testShared(t, cons(), true)
	})
}

func testSequential(t *testing.T, fl flighter) {
	v1, err1, shared1 := fl.Do("t1", func() (any, error) {
		return 0, nil
	})
	v2, err2, shared2 := fl.Do("t1", func() (any, error) {
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

func testShared(t *testing.T, fl flighter, forget bool) {
	synctest.Test(t, func(t *testing.T) {
		ch := make(chan struct{})
		var wg sync.WaitGroup
		defer wg.Wait()
		wg.Go(func() {
			v, err, shared := fl.Do("t1", func() (any, error) {
				<-ch
				return nil, nil
			})
			if v != nil {
				t.Error(1)
			}
			if err != nil {
				t.Error(2)
			}
			if shared == forget {
				t.Error(3)
			}
		})
		synctest.Wait()
		if forget {
			fl.Forget("t1")
		}
		wg.Go(func() {
			v, err, shared := fl.Do("t1", func() (any, error) {
				return nil, nil
			})
			if v != nil {
				t.Error(4)
			}
			if err != nil {
				t.Error(5)
			}
			if shared == forget {
				t.Error(6)
			}
		})
		synctest.Wait()
		close(ch)
	})
}
