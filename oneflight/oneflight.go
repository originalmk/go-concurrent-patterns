package oneflight

import (
	"sync"
)

type Group struct {
	mu   sync.Mutex
	jobs map[string]*job
}

type job struct {
	forgotten bool
	waiters   []chan<- result
}

type result struct {
	res    any
	err    error
	shared bool
}

func (g *Group) Do(key string, f func() (any, error)) (any, error, bool) {
	g.mu.Lock()
	g.ensureFields()

	j, ok := g.jobs[key]
	if !ok || j.forgotten {
		thisJob := &job{
			waiters: make([]chan<- result, 0),
		}
		g.jobs[key] = thisJob
		go func() {
			a, err := f()
			g.mu.Lock()
			defer g.mu.Unlock()
			shared := len(thisJob.waiters) > 1
			for _, waiter := range thisJob.waiters {
				waiter <- result{a, err, shared}
			}
			delete(g.jobs, key)

		}()
	}

	waitCh := make(chan result)
	g.jobs[key].waiters = append(g.jobs[key].waiters, waitCh)
	g.mu.Unlock()

	res := <-waitCh
	return res.res, res.err, res.shared
}

func (g *Group) Forget(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ensureFields()
	j, ok := g.jobs[key]
	if ok {
		j.forgotten = true
	}

}

// Make sure this is called with a lock.
func (g *Group) ensureFields() {
	if g.jobs == nil {
		g.jobs = make(map[string]*job)
	}

}
