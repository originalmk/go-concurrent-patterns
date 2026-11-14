package main

type Tree struct {
	Left  *Tree
	Value int
	Right *Tree
}

func Walk(t *Tree, ch chan int) {
	walk(t, ch)
	close(ch)
}

func walk(t *Tree, ch chan int) {
	if t.Left != nil {
		walk(t.Left, ch)
	}
	ch <- t.Value
	if t.Right != nil {
		walk(t.Right, ch)
	}
}

func Same(t1, t2 *Tree) bool {
	result := true
	ch1 := make(chan int)
	ch2 := make(chan int)
	go Walk(t1, ch1)
	go Walk(t2, ch2)
	for {
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2
		if ok1 != ok2 {
			result = false
		}
		if !ok1 && !ok2 {
			break
		}
		if v1 != v2 {
			result = false
		}
	}
	for range ch1 {
	}
	for range ch2 {
	}
	return result
}
