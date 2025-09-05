package trie

import (
	"sort"
	"strings"
	"sync"

	"github.com/amirzre/autocomplete-system/internal/model"
)

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

// Insert adds a word to the trie or increments its frequency.
func (t *Trie) Insert(word string, frequency int64) {
	if word == "" {
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	word = strings.ToLower(strings.TrimSpace(word))
	node := t.root

	// Traverse/create path for each character
	for _, char := range word {
		if node.children[char] == nil {
			node.children[char] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		node = node.children[char]
	}

	// Mark end of word and set/update frequency
	if !node.isEndOfWord {
		t.size++
	}
	node.isEndOfWord = true
	node.frequency += frequency
	node.word = word
}

// Search finds a word in the trie and returns its frequency.
func (t *Trie) Search(word string) (int64, bool) {
	if word == "" {
		return 0, false
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	word = strings.ToLower(strings.TrimSpace(word))
	node := t.findNode(word)

	if node != nil && node.isEndOfWord {
		return node.frequency, true
	}

	return 0, false
}

// GetSuggestions returns top N suggestions for a given prefix.
func (t *Trie) GetSuggestions(prefix string, limit int) []model.Suggestion {
	if prefix == "" || limit <= 0 {
		return []model.Suggestion{}
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	prefix = strings.ToLower(strings.TrimSpace(prefix))
	prefixNode := t.findNode(prefix)

	if prefixNode == nil {
		return []model.Suggestion{}
	}

	var suggestions []model.Suggestion
	t.collectWords(prefixNode, prefix, &suggestions)

	// Sort by frequency (descending) then alphabetically
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Frequency == suggestions[j].Frequency {
			return suggestions[i].Text < suggestions[j].Text
		}
		return suggestions[i].Frequency > suggestions[j].Frequency
	})

	// Return top N suggestions
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions
}

// GetSize returns the number of unique words in the trie.
func (t *Trie) GetSize() int64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.size
}

// GetTopQueries returns the most frequent queries across the entire trie.
func (t *Trie) GetTopQueries(limit int) []model.Suggestion {
	if limit <= 0 {
		return []model.Suggestion{}
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	var allQueries []model.Suggestion
	t.collectAllWords(t.root, "", &allQueries)

	// Sort by frequency (descending)
	sort.Slice(allQueries, func(i, j int) bool {
		if allQueries[i].Frequency == allQueries[j].Frequency {
			return allQueries[i].Text < allQueries[j].Text
		}
		return allQueries[i].Frequency > allQueries[j].Frequency
	})

	if len(allQueries) > limit {
		allQueries = allQueries[:limit]
	}

	return allQueries
}

// Clear removes all data from the trie.
func (t *Trie) Clear() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.root = &TrieNode{
		children: make(map[rune]*TrieNode),
	}
	t.size = 0
}

// findNode is a helper method to find a node for a given word/prefix.
func (t *Trie) findNode(word string) *TrieNode {
	node := t.root
	for _, char := range word {
		if node.children[char] == nil {
			return nil
		}
		node = node.children[char]
	}
	return node
}

// collectWords recursively collects all words from a subtrie.
func (t *Trie) collectWords(node *TrieNode, currentWord string, suggestions *[]model.Suggestion) {
	if node.isEndOfWord {
		*suggestions = append(*suggestions, model.Suggestion{
			Text:      node.word,
			Frequency: node.frequency,
		})
	}

	for char, childNode := range node.children {
		t.collectWords(childNode, currentWord+string(char), suggestions)
	}
}

// collectAllWords recursively collects all words from the entire trie.
func (t *Trie) collectAllWords(node *TrieNode, currentWord string, queries *[]model.Suggestion) {
	if node.isEndOfWord {
		*queries = append(*queries, model.Suggestion{
			Text:      node.word,
			Frequency: node.frequency,
		})
	}

	for char, childNode := range node.children {
		t.collectAllWords(childNode, currentWord+string(char), queries)
	}
}
