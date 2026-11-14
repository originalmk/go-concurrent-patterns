package main

import "iter"

// Seq[V any]     func(yield func(V) bool)

func (t *Tree) Nodes() iter.Seq[int] {
	return func(yield func(int) bool) {
		ch := make(chan int)
		go Walk(t, ch)
		for n := range ch {
			cont := yield(n)
			if !cont {
				break
			}
		}
		for range ch {
		}
	}
}
