package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

const (
	LINES_PER_PAGE  = 3
	WORDS_PER_LINE  = 8
	MIN_WORD_LENGTH = 3
	MAX_WORD_LENGTH = 8
)

// NOTE: It is an invariant that the word the user is on will always be within
// the last page of the target words.
type config struct {
	linesPerPage      int
	wordsPerLine      int
	minWordLength     int
	maxWordLength     int
	durationInSeconds int
}

type stats struct {
	wpm        int
	accPercent int
}

type metrics struct {
	numCorrectChars   int
	numIncorrectChars int
}

type model struct {
	config        config
	wordGenerator wordGenerator
	currWord      string
	prevWords     []string
	targetWords   []string
	timer         timer.Model
	isDone        bool
	stats         stats
	metrics       metrics
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

func (m model) currWordIndex() int {
	return len(m.prevWords)
}

func (m model) wordIsCorrect(i int) bool {
	if i >= (len(m.prevWords) + 1) {
		return false
	}

	if i == len(m.prevWords) {
		return m.currWord == m.targetWords[i]
	}

	return m.prevWords[i] == m.targetWords[i]
}

func (m model) firstIndexOfPage() int {
	return len(m.targetWords) - m.config.wordsPerLine*m.config.linesPerPage
}

func (m *model) computeStats() {
	var numCharsInCorrectWords = 0

	for i, word := range m.prevWords {
		if m.wordIsCorrect(i) {
			numCharsInCorrectWords += len(word)
		}
	}

	m.stats.wpm = (numCharsInCorrectWords / 5.0) * (60.0 / m.config.durationInSeconds)
	m.stats.accPercent = 100 * m.metrics.numCorrectChars / (m.metrics.numCorrectChars + m.metrics.numIncorrectChars)
}

func initialModel(c config) model {
	generator := newWordGenerator(commonWords, c.minWordLength, c.maxWordLength)

	return model{
		config:        c,
		wordGenerator: generator,
		currWord:      "",
		prevWords:     []string{},
		targetWords:   generator.generate(c.wordsPerLine * c.linesPerPage),
		timer:         timer.New(time.Duration(c.durationInSeconds) * time.Second),
		isDone:        false,
	}
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) DoneUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) lastWordOnCenterLineIndex() int {
	lastTargetWordIndex := len(m.targetWords) - 1
	wordsPerHalfPage := m.config.wordsPerLine * (m.config.linesPerPage / 2)
	return lastTargetWordIndex - wordsPerHalfPage
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.isDone {
		return m.DoneUpdate(msg)
	}

	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.isDone = true
		m.computeStats()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeySpace:
			if m.currWord != "" {
				if m.wordIsCorrect(m.currWordIndex()) {
					m.metrics.numCorrectChars += 1
				} else {
					m.metrics.numIncorrectChars += 1
				}

				m.prevWords = append(m.prevWords, m.currWord)
				m.currWord = ""

				if len(m.prevWords)-1 == m.lastWordOnCenterLineIndex() {
					m.targetWords = append(m.targetWords, m.wordGenerator.generate(m.config.wordsPerLine)...)
				}
			}
		case tea.KeyBackspace:
			if m.currWord != "" {
				m.currWord = removeLastChar(m.currWord)
			} else if len(m.prevWords) > 0 && !m.wordIsCorrect(len(m.prevWords)-1) && len(m.prevWords)-1 >= m.firstIndexOfPage() {
				// If current word is empty and the previous word is not correct and
				// the previous word is on the current page, remove the current word.
				m.currWord = m.prevWord()
				m.prevWords = m.prevWords[:len(m.prevWords)-1]
			}
		default:
			var currTargetWord = m.targetWords[m.currWordIndex()]

			m.currWord += msg.String()

			// If we are still in the word
			if len(m.currWord) <= len(currTargetWord) {
				if m.currWord[len(m.currWord)-1] == currTargetWord[len(m.currWord)-1] {
					m.metrics.numCorrectChars += 1
				} else {
					m.metrics.numIncorrectChars += 1
				}
			} else {
				m.metrics.numIncorrectChars += 1
			}
		}
	}

	return m, nil
}

func (m model) DoneView() string {
	return fmt.Sprintf(
		`typing test completed
wpm: %d
acc: %d/%d = %d%%
press ctrl+c to exit
`,
		m.stats.wpm,
		m.metrics.numCorrectChars,
		m.metrics.numCorrectChars+m.metrics.numIncorrectChars,
		m.stats.accPercent)
}

func (m model) View() (resp string) {
	if m.isDone {
		return m.DoneView()
	}

	resp += TimeRemainingStyle.Render(m.timer.View())
	resp += "\n"

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
	var duration int

	durationFlag := &cli.IntFlag{
		Name:        "duration",
		Usage:       "Specify the duration of the test in seconds",
		Destination: &duration,
		Required:    false,
		Value:       30,
	}

	return &cli.Command{
		Name:  "tt",
		Usage: "Start typing test",
		Flags: []cli.Flag{durationFlag},
		Action: func(ctx *cli.Context) error {
			config := config{
				linesPerPage:      LINES_PER_PAGE,
				wordsPerLine:      WORDS_PER_LINE,
				minWordLength:     MIN_WORD_LENGTH,
				maxWordLength:     MAX_WORD_LENGTH,
				durationInSeconds: duration,
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
