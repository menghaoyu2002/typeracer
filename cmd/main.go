package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	words       []string
	text        string
	charPos     uint
	currentWord uint
	wrongWords  uint
	startTime   time.Time
	hasStarted  bool
	height      int
	width       int
}

type WordList struct {
	Name               string   `json:"name"`
	NoLazyMode         bool     `json:"noLazyMode"`
	OrderedByFrequency bool     `json:"UserPassword"`
	Words              []string `json:"words"`
}

func intiailizeModel() model {
	jsonFile, err := os.ReadFile(path.Join("words/english.json"))
	if err != nil {
		fmt.Println(err)
	}
	var wordlist WordList
	json.Unmarshal(jsonFile, &wordlist)

	generatedWords := [50]string{}
	for i := 0; i < 50; i++ {
		prob := 1 - math.Exp(math.Log(1-rand.Float64()))
		generatedWords[i] = wordlist.Words[uint64(prob*float64(len(wordlist.Words)))]
	}

	model := model{
		words:       generatedWords[:],
		text:        strings.Join(generatedWords[:], " "),
		currentWord: 0,
		charPos:     0,
		wrongWords:  0,
		hasStarted:  false,
	}

	return model
}

func (m model) Init() tea.Cmd {
	return nil
}

func startTimer(m *model) {
	if !m.hasStarted {
		m.startTime = time.Now()
		m.hasStarted = true
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit

		case tea.KeyBackspace.String():
			if m.charPos != 0 {
				if string(m.text[m.charPos-1]) != tea.KeySpace.String() {
					m.charPos--
				}
			}

		default:
			if msg.String() == tea.KeySpace.String() {
				if m.charPos == 0 {
					return m, nil
				}

				m.currentWord++
			}

			if msg.String() == string(m.text[m.charPos]) {

			} else {
			}

			m.charPos++
            startTimer(&m)
		}

	}
	return m, nil
}

func (m model) View() string {
	s := fmt.Sprintf("[%d/%d]\t", m.currentWord, len(m.words))

	wpm := 0
	if m.hasStarted {
		wpm = int(float64(m.currentWord-m.wrongWords) / time.Since(m.startTime).Minutes())
	}

	acc := 100
	if m.currentWord > 0 {
		acc = int((m.currentWord - m.wrongWords) / m.currentWord * 100)
	}
	s += fmt.Sprintf("WPM: %d\t", wpm)
	s += fmt.Sprintf("ACC:  %d", acc)
	s += "\n\n"

	cursor := lipgloss.NewStyle().Blink(true).Underline(true).Bold(true)
	s += m.text[:m.charPos] + cursor.Render(string(m.text[m.charPos])) + m.text[m.charPos+1:]

	style := lipgloss.NewStyle().Width(m.width).PaddingLeft(m.width / 10).PaddingRight(m.width / 10).PaddingTop(m.height / 2).PaddingBottom(m.height / 2)
	return style.Render(s)
}

func main() {
	p := tea.NewProgram(intiailizeModel(), tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("Something went wrong.")
		os.Exit(1)
	}
}
