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
	NUM_LINES       = 3
	WORDS_PER_LINE  = 8
	MAX_WORD_LENGTH = 8
)

type config struct {
	numLines      int
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

func initialModel(c config) model {
	return model{
		config:            c,
		targetStringIndex: 0,
		currWord:          "",
		prevWords:         []string{},
		targetWords:       generateWords(c.wordsPerLine*c.numLines, c.maxWordLength),
	}
}

func (m model) Init() tea.Cmd {
	return nil
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

			// If the user just finished typing the last word on the center line
			// within the last <numLines> lines of target words, generate another
			// line of target words.
			if len(m.prevWords)-1 == len(m.targetWords)-1-m.config.wordsPerLine*(m.config.numLines/2) {
				m.targetWords = append(m.targetWords, generateWords(m.config.wordsPerLine, m.config.maxWordLength)...)
			}
		case tea.KeyBackspace:
			// If there is a single word which is empty, do nothing
			if len(m.prevWords) == 0 && m.currWord == "" {
				return m, nil
			}

			// If current word is empty and the previous word is correct, do nothing
			if len(m.prevWords) > 0 && m.currWord == "" && m.prevWord() == m.targetWords[len(m.prevWords)-1] {
				return m, nil
			}

			// If current word is empty and the previous word is not correct and the previous word is within the last <numLines> lines, remove the current word
			if len(m.prevWords) > 0 && m.currWord == "" && m.prevWord() != m.targetWords[len(m.prevWords)-1] && len(m.prevWords)-1 > len(m.targetWords)-1-m.config.wordsPerLine*m.config.numLines {
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
	// To achieve a scrolling effect where only <numLines> lines of words are
	// shown, only render the last <numLines> lines of target words. For-looping
	// through a ranged slice still starts the index at 0, meaning that here it
	// is no longer the index, but an offset on the start of the range.
	for offset, word := range m.targetWords[len(m.targetWords)-m.config.wordsPerLine*m.config.numLines : len(m.targetWords)] {
		// Calculate the word index using the offset and the starting index.
		indx := len(m.targetWords) - m.config.wordsPerLine*m.config.numLines + offset
		targetWord := []rune(word)
		userWord := []rune(m.getUserWordAtIndex(indx))

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

		// Separate the lines based upon line length.
		if offset > 0 && offset%m.config.wordsPerLine == m.config.wordsPerLine-1 {
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
				numLines:      NUM_LINES,
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
