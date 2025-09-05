package trie

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrieInsert(t *testing.T) {
	trie := New()

	trie.Insert("hello", 1)

	freq, found := trie.Search("hello")
	assert.True(t, found)
	assert.Equal(t, int64(1), freq)
	assert.Equal(t, int64(1), trie.GetSize())
}

func TestTrieInsetMultiple(t *testing.T) {
	trie := New()

	words := []struct {
		word string
		freq int64
	}{
		{"hello", 5},
		{"world", 3},
		{"help", 2},
		{"helicopter", 1},
	}

	for _, w := range words {
		trie.Insert(w.word, w.freq)
	}

	for _, w := range words {
		freq, found := trie.Search(w.word)
		assert.True(t, found, "Word %s should be found", w.word)
		assert.Equal(t, w.freq, freq, "Frequency for %s should match", w.word)
	}

	assert.Equal(t, int64(4), trie.GetSize())
}

func TestTrieIncrementFrequency(t *testing.T) {
	trie := New()

	trie.Insert("test", 1)
	trie.Insert("test", 2)
	trie.Insert("test", 3)

	freq, found := trie.Search("test")
	assert.True(t, found)
	assert.Equal(t, int64(6), freq)
	assert.Equal(t, int64(1), trie.GetSize())
}

func TestTrieSearch(t *testing.T) {
	trie := New()

	freq, found := trie.Search("nonexistent")
	assert.False(t, found)
	assert.Equal(t, int64(0), freq)

	trie.Insert("programming", 10)
	freq, found = trie.Search("programming")
	assert.True(t, found)
	assert.Equal(t, int64(10), freq)

	freq, found = trie.Search("PROGRAMMING")
	assert.True(t, found)
	assert.Equal(t, int64(10), freq)
}

func TestTrieGetSuggestions(t *testing.T) {
	trie := New()

	testData := []struct {
		word string
		freq int64
	}{
		{"javascript", 50},
		{"java", 30},
		{"python", 40},
		{"php", 20},
		{"javascript tutorial", 25},
		{"java programming", 15},
		{"java spring", 10},
	}

	for _, w := range testData {
		trie.Insert(w.word, w.freq)
	}

	suggestions := trie.GetSuggestions("java", 5)
	require.NotEmpty(t, suggestions)

	assert.Equal(t, "javascript", suggestions[0].Text)
	assert.Equal(t, int64(50), suggestions[0].Frequency)

	found := false
	for _, suggestion := range suggestions {
		if suggestion.Text == "java" {
			found = true
			assert.Equal(t, int64(30), suggestion.Frequency)
			break
		}
	}
	assert.True(t, found, "Should include exact match 'java'")
}

func TestTrieGetSuggestionsLimit(t *testing.T) {
	trie := New()

	words := []string{"test1", "test2", "test3", "test4", "test5", "test6"}
	for _, word := range words {
		trie.Insert(word, 1)
	}

	suggestions := trie.GetSuggestions("test", 3)
	assert.Equal(t, 3, len(suggestions))
}

func TestTrieGetSuggestionsEmptyPrefix(t *testing.T) {
	trie := New()
	trie.Insert("test", 1)

	suggestions := trie.GetSuggestions("", 5)
	assert.Empty(t, suggestions)
}

func TestTrieGetSuggestionsNoMatches(t *testing.T) {
	trie := New()
	trie.Insert("hello", 1)

	suggestions := trie.GetSuggestions("xyz", 5)
	assert.Empty(t, suggestions)
}

func TestTrieGetTopQueries(t *testing.T) {
	trie := New()

	testData := []struct {
		word string
		freq int64
	}{
		{"popular", 100},
		{"medium", 50},
		{"low", 10},
		{"lowest", 5},
		{"highest", 150},
	}

	for _, data := range testData {
		trie.Insert(data.word, data.freq)
	}

	topQueries := trie.GetTopQueries(3)
	require.Equal(t, 3, len(topQueries))

	assert.Equal(t, "highest", topQueries[0].Text)
	assert.Equal(t, int64(150), topQueries[0].Frequency)

	assert.Equal(t, "popular", topQueries[1].Text)
	assert.Equal(t, int64(100), topQueries[1].Frequency)

	assert.Equal(t, "medium", topQueries[2].Text)
	assert.Equal(t, int64(50), topQueries[2].Frequency)
}

func TestTrieClear(t *testing.T) {
	trie := New()

	trie.Insert("test1", 1)
	trie.Insert("test2", 2)
	assert.Equal(t, int64(2), trie.GetSize())

	trie.Clear()
	assert.Equal(t, int64(0), trie.GetSize())

	freq, found := trie.Search("test1")
	assert.False(t, found)
	assert.Equal(t, int64(0), freq)
}

func TestTrieCaseSensitivity(t *testing.T) {
	trie := New()

	trie.Insert("Hello", 1)
	trie.Insert("HELLO", 2)
	trie.Insert("hello", 3)

	freq, found := trie.Search("hello")
	assert.True(t, found)
	assert.Equal(t, int64(6), freq)
	assert.Equal(t, int64(1), trie.GetSize())
}

func TestTrieSpecialCharacters(t *testing.T) {
	trie := New()

	words := []string{
		"hello world",
		"test-case",
		"file.txt",
		"user@domain.com",
	}

	for _, word := range words {
		trie.Insert(word, 1)
	}

	for _, word := range words {
		freq, found := trie.Search(word)
		assert.True(t, found, "Should find word: %s", word)
		assert.Equal(t, int64(1), freq)
	}
}

func TestTrieEmptyString(t *testing.T) {
	trie := New()

	trie.Insert("", 5)
	assert.Equal(t, int64(0), trie.GetSize())

	freq, found := trie.Search("")
	assert.False(t, found)
	assert.Equal(t, int64(0), freq)
}
