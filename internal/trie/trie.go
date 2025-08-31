package trie

import "sync"

// TrieNode represents a node in the trie.
type TrieNode struct {
	children    map[rune]*TrieNode
	isEndOfWord bool
	frequency   int64
	word        string
}

// Trie represents the main trie data structure with thread safety.
type Trie struct {
	root  *TrieNode
	mutex sync.RWMutex
	size  int64
}

// New creates a new Trie instance.
func New() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}
