package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

var (
	targetWords = []string{"kartik", "aaron", "josh"}
)

type model struct {
	targetStringIndex int
	currWord          string
	prevWords         []string
	targetWords       []string
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
		targetWords:       targetWords,
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

func (m model) View() string {
	return m.RenderUserInput()
}

func (m model) RenderUserInput() string {
	var sb strings.Builder

	for indx, word := range m.targetWords {
		targetWord := []rune(word)
		userWord := []rune(m.getUserWordAtIndex(indx))

		for i := 0; i < int(math.Min(float64(len(targetWord)), float64(len(userWord)))); i++ {
			textStyle := CorrectTextStyle
			if targetWord[i] != userWord[i] {
				textStyle = WrongTextStyle
			}
			sb.WriteString(textStyle.Render(string(targetWord[i])))
		}

		if len(userWord) > len(targetWord) {
			sb.WriteString(WrongTextStyle.Render(string(userWord[len(targetWord):])))
		} else {
			sb.WriteString(UnreachedTextStyle.Render(string(targetWord[len(userWord):])))
		}
		sb.WriteString(" ")
	}

	return sb.String()
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
