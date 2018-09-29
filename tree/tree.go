// Package tree defines a generic binar tree structure
package tree

// Tree defines a binary tree structure
type Tree struct {
	root    *node
	compare Comparable
}

type node struct {
	left  *node
	right *node
	value interface{}
}

// Comparable defines a function signature for comparing two objects
// this will be used by the binary tree for insertions
// It returns true if ob1 is less than ob2
// and false otherwise
type Comparable func(ob1 interface{}, ob2 interface{}) int

// NewTree creates and returns a new binary tree
// compareFun is a function used to determine if a new value being inserted
// is less than the current value at a specific node
func NewTree(compareFun Comparable) *Tree {
	return &Tree{compare: compareFun}
}

func (t *Tree) insert(n *node, val interface{}) *node {
	if n == nil {
		return &node{value: val}
	}
	r := t.compare(val, n.value)
	if r < 0 {
		n.left = t.insert(n.left, val)
	} else if r > 0 {
		n.right = t.insert(n.right, val)
	} else {
		return n
	}
	return n
}

type orderedArr struct {
	arr []interface{}
}

func (t *Tree) inOrderArr(n *node, arr *orderedArr) {
	if n != nil {
		t.inOrderArr(n.left, arr)
		arr.arr = append(arr.arr, n.value)
		t.inOrderArr(n.right, arr)
	}
}

// Insert adds a value to the tree. Duplicates will not be added.
func (t *Tree) Insert(val interface{}) {
	t.root = t.insert(t.root, val)
}

// InOrderArr builds a sorted array from the values in the tree
func (t *Tree) InOrderArr() []interface{} {
	inOrderArray := &orderedArr{}
	t.inOrderArr(t.root, inOrderArray)
	return inOrderArray.arr
}
