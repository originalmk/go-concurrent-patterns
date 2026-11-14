package main

import (
	"slices"
	"testing"
)

func TestSame(t *testing.T) {
	t1 := firstTree()
	t2 := secondTree()
	same := Same(t1, t2)
	if !same {
		t.Fatal("expected the trees to be the same")
	}
	t2.Left.Left.Right.Value = 42
	same = Same(t1, t2)
	if same {
		t.Fatal("trees should not be the same after mutation")
	}
}

func TestWalk(t *testing.T) {
	expectedElems := []int{1, 1, 2, 3, 5, 8, 13}
	t.Run("first tree walk", testWalk(firstTree(), expectedElems))
	t.Run("second tree walk", testWalk(secondTree(), expectedElems))
}

func testWalk(tree *Tree, expectedElems []int) func(t *testing.T) {
	return func(t *testing.T) {
		ch := make(chan int)
		go Walk(tree, ch)
		gotElems := make([]int, 0)
		for elem := range ch {
			gotElems = append(gotElems, elem)
		}
		if !slices.Equal(gotElems, expectedElems) {
			t.Errorf("expected %+v, got %+v", expectedElems, gotElems)
		}
	}
}

func firstTree() *Tree {
	return &Tree{
		Value: 3,
		Left: &Tree{
			Value: 1,
			Left:  &Tree{Value: 1},
			Right: &Tree{Value: 2},
		},
		Right: &Tree{
			Value: 8,
			Left:  &Tree{Value: 5},
			Right: &Tree{Value: 13},
		},
	}
}

func secondTree() *Tree {
	return &Tree{
		Value: 8,
		Left: &Tree{
			Value: 3,
			Left: &Tree{
				Value: 1,
				Left:  &Tree{Value: 1},
				Right: &Tree{Value: 2},
			},
			Right: &Tree{Value: 5},
		},
		Right: &Tree{Value: 13},
	}
}
