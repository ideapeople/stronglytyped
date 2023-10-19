package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tjarratt/babble"
	"github.com/urfave/cli/v2"
)

const (
	LINES_PER_PAGE  = 3
	WORDS_PER_LINE  = 8
	MAX_WORD_LENGTH = 8
)

// NOTE: It is an invariant that the word the user is on will always be within
// the last page of the target words.
type config struct {
	linesPerPage  int
	wordsPerLine  int
	maxWordLength int
}

type model struct {
	config            config
	targetStringIndex int
	currWord          string
	prevWords         []string
	targetWords       []string
}

func generateWords(numWords int, maxWordLength int) []string {
	babbler := babble.NewBabbler()
	babbler.Count = numWords
	babbler.Separator = " "
	babbler.Words = fold(babbler.Words, []string{}, func(word string, acc []string) []string {
		// Ignore words longer than the max word length.
		if len(word) > maxWordLength {
			return acc
		}

		// Lowercase added words.
		return append(acc, strings.ToLower(word))
	})

	return strings.Split(babbler.Babble(), babbler.Separator)
}

func (m model) getUserWordAtIndex(index int) string {
	if len(m.prevWords) < index {
		return ""
	}
	if len(m.prevWords) == index {
		return m.currWord
	}
	return m.prevWords[index]
}

func (m model) prevWord() string {
	return m.prevWords[len(m.prevWords)-1]
}

func (m model) wordIsCorrect(i int) bool {
	if i >= len(m.targetWords) {
		return false
	}

	if i == len(m.targetWords)-1 {
		return m.currWord == m.targetWords[i]
	}

	return m.prevWords[i] == m.targetWords[i]
}

func (m model) firstIndexOfPage() int {
	return len(m.targetWords) - m.config.wordsPerLine*m.config.linesPerPage
}

func initialModel(c config) model {
	return model{
		config:            c,
		targetStringIndex: 0,
		currWord:          "",
		prevWords:         []string{},
		targetWords:       generateWords(c.wordsPerLine*c.linesPerPage, c.maxWordLength),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) lastWordOnCenterLineIndex() int {
	lastTargetWordIndex := len(m.targetWords) - 1
	wordsPerHalfPage := m.config.wordsPerLine * (m.config.linesPerPage / 2)
	return lastTargetWordIndex - wordsPerHalfPage
}

func (m model) generateWords() []string {
	return generateWords(m.config.wordsPerLine, m.config.maxWordLength)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeySpace:
			if m.currWord == "" {
				return m, nil
			}

			m.prevWords = append(m.prevWords, m.currWord)
			m.currWord = ""

			if len(m.prevWords)-1 == m.lastWordOnCenterLineIndex() {
				m.targetWords = append(m.targetWords, m.generateWords()...)
			}
		case tea.KeyBackspace:
			// If there is a single word which is empty, do nothing
			if len(m.prevWords) == 0 && m.currWord == "" {
				return m, nil
			}

			// If current word is empty and the previous word is correct, do nothing
			if len(m.prevWords) > 0 && m.currWord == "" && m.wordIsCorrect(len(m.prevWords)-1) {
				return m, nil
			}

			// If current word is empty and the previous word is not correct and the
			// previous word is on the current page, remove the current word
			prevWordIsOnPage := len(m.prevWords)-1 >= m.firstIndexOfPage()
			if len(m.prevWords) > 0 && m.currWord == "" && !m.wordIsCorrect(len(m.prevWords)-1) && prevWordIsOnPage {
				m.currWord = m.prevWord()
				m.prevWords = m.prevWords[:len(m.prevWords)-1]
				return m, nil
			}

			// Else remove the last character of the current word
			m.currWord = removeLastChar(m.currWord)
		default:
			m.currWord += msg.String()
		}
	}

	return m, nil
}

func (m model) View() (resp string) {
	// To achieve a scrolling effect where only <linesPerPage> lines of words are
	// shown, only render the last <linesPerPage> lines of target words.
	// For-looping through a ranged slice still starts the index at 0, meaning
	// that here it is no longer the index, but an offset on the range start.
	for offset, word := range m.targetWords[m.firstIndexOfPage():len(m.targetWords)] {
		var (
			indx       = m.firstIndexOfPage() + offset
			targetWord = []rune(word)
			userWord   = []rune(m.getUserWordAtIndex(indx))
		)

		for i := 0; i < min(len(targetWord), len(userWord)); i++ {
			textStyle := CorrectTextStyle
			if targetWord[i] != userWord[i] {
				textStyle = WrongTextStyle
			}
			resp += textStyle.Render(string(targetWord[i]))
		}

		// determines when to print cursor
		var cursor string
		if len(m.prevWords) == indx {
			cursor = CursorTextStyle.Render("|")
		}

		if len(userWord) > len(targetWord) {
			resp += OvertypedTextStyle.Render(string(userWord[len(targetWord):])) + cursor
		} else {
			resp += cursor + UnreachedTextStyle.Render(string(targetWord[len(userWord):]))
		}

		indexAtEndOfLine := offset > 0 && offset%m.config.wordsPerLine == m.config.wordsPerLine-1
		if indexAtEndOfLine {
			resp += "\n"
		} else if indx < len(m.targetWords)-1 {
			resp += " "
		}
	}

	return
}

func removeLastChar(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s // or return an error if you prefer
	}
	return string(runes[:len(runes)-1])
}

func main() {
	_, rootCancel := signal.NotifyContext(context.Background(), os.Interrupt)

	app := &cli.App{
		Name:  "stronglytyped",
		Usage: "Typing test",
		Authors: []*cli.Author{
			{
				Name: "ideapeople",
			},
		},
		EnableBashCompletion: true,
		BashComplete:         cli.DefaultAppComplete,
		Commands: []*cli.Command{
			startTest(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		rootCancel()
		log.Fatal(err)
	}

	rootCancel()
}

func startTest() *cli.Command {
	var length int

	lengthFlag := &cli.IntFlag{
		Name:        "length",
		Usage:       "Specify the length of the test in seconds",
		Destination: &length,
		Required:    false,
		Value:       30,
	}

	return &cli.Command{
		Name:  "tt",
		Usage: "Start typing test",
		Flags: []cli.Flag{lengthFlag},
		Action: func(ctx *cli.Context) error {
			config := config{
				linesPerPage:  LINES_PER_PAGE,
				wordsPerLine:  WORDS_PER_LINE,
				maxWordLength: MAX_WORD_LENGTH,
			}

			p := tea.NewProgram(initialModel(config))
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}

			return nil
		},
	}
}
