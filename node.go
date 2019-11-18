package bplustree

const (
	Block = 4096
	MaxKV = Block / 50
	MaxKC = (Block - 46) / 50
)

type node interface {
	find(key string) (int, bool)
	parent() *interiorNode
	setParent(*interiorNode)
	full() bool
}
