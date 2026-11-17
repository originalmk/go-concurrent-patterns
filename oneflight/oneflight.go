package oneflight

import "sync"

type Group struct {
	mu      sync.Mutex
	waiters map[string][]chan<- result
	jobs    map[string]struct{}
}

type result struct {
	res    any
	err    error
	shared bool
}

func (g *Group) Do(key string, f func() (any, error)) (any, error, bool) {
	g.mu.Lock()

	if g.waiters == nil {
		g.waiters = make(map[string][]chan<- result)
	}
	if g.jobs == nil {
		g.jobs = make(map[string]struct{})
	}

	_, ok := g.jobs[key]
	if !ok {
		g.jobs[key] = struct{}{}
		g.waiters[key] = make([]chan<- result, 0)
		go func() {
			a, err := f()
			g.mu.Lock()
			defer g.mu.Unlock()
			shared := len(g.waiters[key]) > 1
			for _, waiter := range g.waiters[key] {
				waiter <- result{a, err, shared}
			}
		}()
	}

	waitCh := make(chan result)
	g.waiters[key] = append(g.waiters[key], waitCh)
	g.mu.Unlock()

	res := <-waitCh
	return res.res, res.err, res.shared
}
