package main

import (
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/tjarratt/babble"
)

type wordGenerator struct {
	// These properties are only meant to be used internally to track when either
	// the min or max word length properties changed so that the words array can
	// be refreshed from the original words and memoized.
	prevMinWordLength int
	prevMaxWordLength int

	minWordLength int
	maxWordLength int
	originalWords []string
	words         []string
}

func newWordGenerator(words []string, minWordLength int, maxWordLength int) wordGenerator {
	return wordGenerator{
		prevMinWordLength: minWordLength,
		prevMaxWordLength: maxWordLength,

		minWordLength: minWordLength,
		maxWordLength: maxWordLength,
		originalWords: words,
		words:         filterAndLowercaseWords(words, minWordLength, maxWordLength),
	}
}

func filterAndLowercaseWords(words []string, minWordLength int, maxWordLength int) []string {
	return filter(words, func(word string) bool {
		return minWordLength <= len(word) && len(word) <= maxWordLength
	})
}

func (w wordGenerator) generate(n int) (res []string) {
	// If either bound has changed, refresh the words.
	if w.prevMinWordLength != w.minWordLength || w.prevMaxWordLength != w.maxWordLength {
		w.prevMinWordLength = w.minWordLength
		w.prevMaxWordLength = w.maxWordLength
		w.words = filterAndLowercaseWords(w.originalWords, w.minWordLength, w.maxWordLength)
	}

	for i := 0; i < n; i++ {
		res = append(res, w.words[rand.Int()%len(w.words)])
	}

	return
}

func getCommonWords() (words []string) {
	file, err := os.Open("./common_words.txt")
	if err != nil {
		panic(err)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	words = strings.Split(string(bytes), "\n")
	return
}

var (
	commonWords = getCommonWords()
	unixWords   = babble.NewBabbler().Words
)
