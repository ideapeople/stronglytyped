package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	"log"
	"os"
	"os/signal"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

var TARGET_WORDS = []string{"kartik", "aaron", "josh"}

type model struct {
	textInput         textinput.Model
	targetStringIndex int
	currWord          string
	prevWords         []string
}

func (m model) prevWord() string {
	return m.prevWords[len(m.prevWords)-1]
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	return model{
		textInput:         ti,
		targetStringIndex: 0,
		currWord:          "",
		prevWords:         []string{},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeySpace:
			if m.currWord == "" {
				return m, nil
			}

			if len(m.prevWords) == len(TARGET_WORDS)-1 {
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
			if len(m.prevWords) > 0 && m.currWord == "" && m.prevWord() == TARGET_WORDS[len(m.prevWords)-1] {
				return m, nil
			}

			// If current word is empty and the previous word is not correct, remove the current word
			if len(m.prevWords) > 0 && m.currWord == "" && m.prevWord() != TARGET_WORDS[len(m.prevWords)-1] {
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
	result := strings.Join(TARGET_WORDS, " ") + "\n" + strings.Join(m.prevWords, " ")

	if len(m.prevWords) > 0 {
		result += " "
	}

	result += m.currWord

	return result
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
