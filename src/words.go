package main

import (
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/tjarratt/babble"
)

type wordGenerator struct {
	minWordLength int
	maxWordLength int
	originalWords []string
	words         []string
}

func newWordGenerator(words []string, minWordLength int, maxWordLength int) wordGenerator {
	return wordGenerator{
		minWordLength: minWordLength,
		maxWordLength: maxWordLength,
		originalWords: words,
		words:         filterAndLowercaseWords(words, minWordLength, maxWordLength),
	}
}

func filterAndLowercaseWords(words []string, minWordLength int, maxWordLength int) []string {
	return fold(words, []string{}, func(acc []string, word string) []string {
		if len(word) < minWordLength && maxWordLength < len(word) {
			return acc
		}

		return append(acc, strings.ToLower(word))
	})
}

func (w wordGenerator) setBounds(minWordLength int, maxWordLength int) {
	w.minWordLength = minWordLength
	w.maxWordLength = maxWordLength
	w.words = filterAndLowercaseWords(w.originalWords, w.minWordLength, w.maxWordLength)
}

func (w wordGenerator) setWords(words []string) {
	w.originalWords = words
	w.words = filterAndLowercaseWords(w.originalWords, w.minWordLength, w.maxWordLength)
}

func (w wordGenerator) generate(n int) (res []string) {
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
