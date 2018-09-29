package tree

import (
	"testing"
)

func compare(ob1 interface{}, ob2 interface{}) int {
	if ob1.(int) < ob2.(int) {
		return -1
	} else if ob1.(int) > ob2.(int) {
		return 1
	}
	return 0
}

func buildTree() *Tree {
	// Tree will be:
	//      6
	//    /  \
	//   3    7
	//  / \
	// 2   4
	//      \
	//       5
	tree := NewTree(compare)
	tree.Insert(6)
	tree.Insert(3)
	tree.Insert(7)
	tree.Insert(4)
	tree.Insert(2)
	tree.Insert(5)
	// no duplicates
	tree.Insert(6)
	return tree
}

// TestNewTree tests that the NewTree function returns an empty tree structure.
func TestNewTree(t *testing.T) {
	tree := NewTree(nil)
	if tree == nil {
		t.Error("NewTree returned nil")
	}
	if tree.compare != nil {
		t.Error("Tree's compare wasn't nil")
	}
	tree = NewTree(compare)
	if tree.compare(4, 10) != -1 {
		t.Error("Tree's compare function evaluated ints incorrectly")
	}
	if tree.root != nil {
		t.Error("Tree's value should have been nil")
	}
}

// TestInsert uses the Tree's insert method to build a tree.
// It makes sure the tree has the correct branches after calling
// insert multiple times.
func TestInsert(t *testing.T) {
	tree := buildTree()
	if tree.root.value != 6 {
		t.Error("Root node's value should have been 5")
	}
	if tree.root.left.value != 3 {
		t.Error("Node's values should have been 4")
	}
	if tree.root.right.value != 7 {
		t.Error("Node's value should have been 7")
	}
	if tree.root.left.left.value != 2 {
		t.Error("Node's value should have been 3")
	}
	if tree.root.left.right.value != 4 {
		t.Error("Node's value should have been 4")
	}
	if tree.root.left.right.right.value != 5 {
		t.Error("Node's value should have been 5")
	}
}

// TestInOrder builds a tree and tests that the InOrderArr creates a sorted
// array based on the values in the tree.
func TestInOrder(t *testing.T) {
	tr := buildTree()
	valArr := tr.InOrderArr()
	for i, v := range valArr {
		if v != (i + 2) {
			t.Fatalf("Value was supposed to be: %d but was %d", i+2, v)
		}
	}
}
