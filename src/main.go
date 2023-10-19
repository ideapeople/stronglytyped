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

const WORD_PAGE_LENGTH = 10
const MAX_WORD_LENGTH = 8

type model struct {
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
		if len(word) > MAX_WORD_LENGTH {
			return acc
		}

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

func initialModel() model {
	return model{
		targetStringIndex: 0,
		currWord:          "",
		prevWords:         []string{},
		targetWords:       generateWords(WORD_PAGE_LENGTH, MAX_WORD_LENGTH),
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

			if len(m.prevWords) == len(m.targetWords)-1 {
				return m, tea.Quit
			}

			m.prevWords = append(m.prevWords, m.currWord)
			m.currWord = ""
		case tea.KeyBackspace:
			// If there is a single word which is empty, do nothing
			if len(m.prevWords) == 0 && m.currWord == "" {
				return m, nil
			}

			// If current word is empty and the previous word is correct, do nothing
			if len(m.prevWords) > 0 && m.currWord == "" && m.prevWord() == m.targetWords[len(m.prevWords)-1] {
				return m, nil
			}

			// If current word is empty and the previous word is not correct, remove the current word
			if len(m.prevWords) > 0 && m.currWord == "" && m.prevWord() != m.targetWords[len(m.prevWords)-1] {
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
	for indx, word := range m.targetWords {
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

		if indx < len(m.targetWords)-1 {
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
			p := tea.NewProgram(initialModel())
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}

			return nil
		},
	}
}
