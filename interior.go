package bplustree

import (
	"sort"
)

type kc struct {
	key   string
	child node
}

// one empty slot for split
type kcs [MaxKC + 1]kc

func (a *kcs) Len() int { return len(a) }

func (a *kcs) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a *kcs) Less(i, j int) bool {
	if a[i].key == "" {
		return false
	}

	if a[j].key == "" {
		return true
	}

	return a[i].key < a[j].key
}

type interiorNode struct {
	kcs   kcs
	count int
	p     *interiorNode
}


func newInteriorNode(p *interiorNode, largestChild node) *interiorNode {
	i := &interiorNode{
		p:     p,
		count: 1,
	}

	if largestChild != nil {
		i.kcs[0].child = largestChild
	}
	return i
}

func (in *interiorNode) find(key string) (int, bool) {
	c := func(i int) bool { return in.kcs[i].key > key }

	i := sort.Search(in.count-1, c)

	return i, true
}

func (in *interiorNode) full() bool { return in.count == MaxKC }

func (in *interiorNode) parent() *interiorNode { return in.p }

func (in *interiorNode) setParent(p *interiorNode) { in.p = p }

func (in *interiorNode) insert(key string, child node) (string, *interiorNode, bool) {
	i, _ := in.find(key)

	if !in.full() {
		copy(in.kcs[i+1:], in.kcs[i:in.count])

		in.kcs[i].key = key
		in.kcs[i].child = child
		child.setParent(in)

		in.count++
		return "", nil, false
	}

	// insert the new node into the empty slot
	in.kcs[MaxKC].key = key
	in.kcs[MaxKC].child = child
	child.setParent(in)

	next, midKey := in.split()

	return midKey, next, true
}

func (in *interiorNode) split() (*interiorNode, string) {
	sort.Sort(&in.kcs)

	// get the mid info
	midIndex := MaxKC / 2
	midChild := in.kcs[midIndex].child
	midKey := in.kcs[midIndex].key

	// create the split node with out a parent
	next := newInteriorNode(nil, nil)
	copy(next.kcs[0:], in.kcs[midIndex+1:])
	next.count = MaxKC - midIndex
	// update parent
	for i := 0; i < next.count; i++ {
		next.kcs[i].child.setParent(next)
	}

	// modify the original node
	in.count = midIndex + 1
	in.kcs[in.count-1].key = ""
	in.kcs[in.count-1].child = midChild
	midChild.setParent(in)

	return next, midKey
}
