package bplustree

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"zly.ecnu.edu.cn/bplustree/ipfs"
)

type BTree struct {
	root     *interiorNode
	first    *leafNode
	leaf     int
	interior int
	height   int
}

func newBTree() *BTree {
	leaf := newLeafNode(nil)
	r := newInteriorNode(nil, leaf)
	leaf.p = r
	return &BTree{
		root:     r,
		first:    leaf,
		leaf:     1,
		interior: 1,
		height:   2,
	}
}



// first returns the first leafNode
func (bt *BTree) First() *leafNode {
	return bt.first
}

// insert inserts a (key, value) into the B+ tree
func (bt *BTree) Insert(key string, value string) {
	_, oldIndex, leaf := search(bt.root, key)
	p := leaf.parent()

	mid, bump := leaf.insert(key, value)
	if !bump {
		return
	}

	var midNode node
	midNode = leaf

	p.kcs[oldIndex].child = leaf.next
	leaf.next.setParent(p)

	interior, interiorP := p, p.parent()

	for {
		var oldIndex int
		var newNode *interiorNode

		isRoot := interiorP == nil

		if !isRoot {
			oldIndex, _ = interiorP.find(key)
		}

		mid, newNode, bump = interior.insert(mid, midNode)
		if !bump {
			return
		}

		if !isRoot {
			interiorP.kcs[oldIndex].child = newNode
			newNode.setParent(interiorP)

			midNode = interior
		} else {
			bt.root = newInteriorNode(nil, newNode)
			newNode.setParent(bt.root)

			bt.root.insert(mid, interior)
			return
		}

		interior, interiorP = interiorP, interior.parent()
	}
}

// Search searches the key in B+ tree
// If the key exists, it returns the value of key and true
// If the key does not exist, it returns an empty string and false
func (bt *BTree) Search(key string) (string, bool) {
	kv, _, _ := search(bt.root, key)
	if kv == nil {
		return "", false
	}
	return kv.value, true
}

func search(n node, key string) (*kv, int, *leafNode) {
	curr := n
	oldIndex := -1

	for {
		switch t := curr.(type) {
		case *leafNode:
			i, ok := t.find(key)
			if !ok {
				return nil, oldIndex, t
			}
			return &t.kvs[i], oldIndex, t
		case *interiorNode:
			i, _ := t.find(key)
			curr = t.kcs[i].child
			oldIndex = i
		default:
			panic("")
		}
	}
}

func (bt *BTree) Traversal() {
	root := postOrderTraversal(bt.root)
	//cids := make([]string, 0)
	//
	//for index, n := range bt.root.kcs {
	//	if index >= bt.root.count {
	//		break
	//	}
	//	cids = append(cids, postOrderTraversal(n.child))
	//}
	//fmt.Println(cids)
	fmt.Println(root)

}

func postOrderTraversal(n node) string {
	result := make([]string, 0)
	res := ""
	switch t := n.(type) {
	case *interiorNode:
		hash := ""
		keys := make([]string, 0)
		for index, kc := range t.kcs {
			if index >= t.count {
				break
			}

			hash = postOrderTraversal(kc.child)
			result = append(result, hash)
		}

		for index, kc := range t.kcs {
			if index >= t.count - 1 {
				break
			}
			keys = append(keys, kc.key)
		}


		upiNode := &uploadNode {
			Type: 1,
			Keys: keys,
			Cids: result,
		}

		in, _ := json.Marshal(upiNode)
		res = string(in)

	case *leafNode:
		var sb strings.Builder
		keys := make([]string, 0)
		value := make([]string, 0)
		length := 0
		start := 0
		for i := 0; i < t.count; i++ {
			keys = append(keys, t.kvs[i].key)
			curLength := len(t.kvs[i].key) +len(t.kvs[i].value) + 1
			if length + curLength > Block{
				cid := ipfs.UploadIndex(sb.String())
				for j := start; j < i; j++ {
					value = append(value, cid)
				}
				sb.Reset()
				length = 0
				start = i
			}
			length += curLength
			sb.WriteString(t.kvs[i].key)
			sb.WriteString(" ")
			sb.WriteString(t.kvs[i].value)
		}

		cid := ipfs.UploadIndex(sb.String())
		for j := start; j < t.count; j++ {
			value = append(value, cid)
		}

		uplNode := &uploadNode{
			Type: 0,
			Keys: keys,
			Cids: value,
		}

		in, _ := json.Marshal(uplNode)
		res = string(in)
	}

	return ipfs.UploadIndex(res)
}

func keyWordQuery(cid string, keyWord string) string {
	curr := cid

	content := ipfs.CatIndex(curr)
	n := &uploadNode{}
	err := json.Unmarshal([]byte(content), &n)

	if err != nil {
		log.Printf("Find index error: unmarshal error: %v", err)
	}

	for {
		switch n.Type {
		case 0 :
			i := sort.Search(len(n.Keys), func(i int) bool {
				return keyWord <= n.Keys[i]
			})


			if i < len(n.Keys) && n.Keys[i] == keyWord {
				content := ipfs.CatIndex(n.Cids[i])
				contents := strings.Split(content, "\n")
				for _, w := range contents {
					results := strings.Split(w, " ")
					if results[0] == keyWord {
						return results[1]
					}
				}
			}

			return ""
		case 1:
			c := func(i int) bool {
				return n.Keys[i] >= keyWord
			}

			i := sort.Search(len(n.Keys), c)

			if i == -1 {
				content = ipfs.CatIndex(n.Cids[i+1])
			} else {
				content = ipfs.CatIndex(n.Cids[i])
			}

			n = &uploadNode{}
			err := json.Unmarshal([]byte(content), &n)

			if err != nil {
				log.Printf("Find index error: unmarshal error: %v", err)
			}

		default:
			panic("")
		}
	}
}


