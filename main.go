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
	words         []string
	text          string
	formattedText []string
	charPos       int
	currentWord   int
	wrongWords    []bool
	currMistakes  []bool
	wordPos       int
	keystrokes    int
	mistakes      int
	startTime     time.Time
	duration      time.Duration
	hasStarted    bool
	hasEnded      bool
	height        int
	width         int
}

type WordList struct {
	Name               string   `json:"name"`
	NoLazyMode         bool     `json:"noLazyMode"`
	OrderedByFrequency bool     `json:"UserPassword"`
	Words              []string `json:"words"`
}

func generateWords() []string {
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

	return generatedWords[:]
}
func intiailizeModel() model {
	generatedWords := generateWords()
	text := strings.Join(generatedWords[:], " ") + " "
	model := model{
		words:         generatedWords[:],
		text:          text,
		formattedText: make([]string, len(text)),
		currentWord:   0,
		charPos:       0,
		wordPos:       0,
		wrongWords:    make([]bool, len(generatedWords)),
		currMistakes:  make([]bool, len(generatedWords[0])+1),
		mistakes:      0,
		keystrokes:    0,
		hasStarted:    false,
		hasEnded:      false,
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
	correctColor := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	wrongColor := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit

		// case "ctrl+r":
		//           newModel := intiailizeModel()
		//           newModel.height = m.height
		//           newModel.width = m.width
		//           m = newModel

		case tea.KeyBackspace.String():
			if m.charPos != 0 && !m.hasEnded {
				if string(m.text[m.charPos-1]) != tea.KeySpace.String() {
					m.charPos--
					m.wordPos--
					m.currMistakes[m.wordPos] = false
					m.keystrokes++
					m.formattedText[m.charPos] = ""
				}
			}

		default:
			if m.hasEnded || (string(m.text[m.charPos]) == tea.KeySpace.String() && msg.String() != tea.KeySpace.String()) {
				return m, nil
			}

			if msg.String() == tea.KeySpace.String() {
				if m.charPos == 0 {
					return m, nil
				}

				for string(m.text[m.charPos]) != tea.KeySpace.String() && m.charPos < len(m.text)-2 {
					m.formattedText[m.charPos] = wrongColor.Render(string(m.text[m.charPos]))
					m.currMistakes[m.wordPos] = true
					m.charPos++
					m.wordPos++
				}

				if sum(m.currMistakes) > 0 {
					m.wrongWords[m.currentWord] = true
				}

				m.currentWord++
				if m.currentWord < len(m.words) {
					m.currMistakes = make([]bool, len(m.words[m.currentWord])+1)
				}
				m.wordPos = 0
			}

			if msg.String() == string(m.text[m.charPos]) {
				m.formattedText[m.charPos] = correctColor.Render(string(m.text[m.charPos]))
			} else {
				m.mistakes++
				m.currMistakes[m.wordPos] = true
				m.formattedText[m.charPos] += wrongColor.Render(string(m.text[m.charPos]))
			}

			m.charPos++
			m.wordPos++
			m.keystrokes++

			startTimer(&m)
			if m.charPos == len(m.text)-1 {
				m.hasEnded = true
				m.duration = time.Since(m.startTime)
				m.currentWord = len(m.words)
			}

		}

	}
	return m, nil
}

func sum(arr []bool) int {
	sum := 0
	for _, truth := range arr {
		if truth {
			sum++
		}
	}
	return sum
}

func (m model) View() string {
	s := fmt.Sprintf("[%d/%d]\t", m.currentWord, len(m.words))

	if m.hasEnded {
		wpm := int(float64(m.currentWord-sum(m.wrongWords)) / m.duration.Minutes())
		acc := int(float64(m.keystrokes-m.mistakes) / float64(m.keystrokes) * 100)
		s += fmt.Sprintf("WPM: %d\t", wpm)
		s += fmt.Sprintf("ACC:  %d", acc) + "%"
	}

	s += "\n\n"

	cursor := lipgloss.
		NewStyle().
		Background(lipgloss.AdaptiveColor{Light: "15", Dark: "0"})

	faint := lipgloss.NewStyle().Faint(true)
	s += strings.Join(m.formattedText, "") + cursor.Render(string(m.text[m.charPos])) + faint.Render(m.text[m.charPos+1:])

	// s += faint.Render("\n\n\n\nCtrl+C or Ctrl+D to quit")

	layout := lipgloss.NewStyle().
		Width(m.width).
		PaddingLeft(m.width / 10).
		PaddingRight(m.width / 10).
		PaddingTop(m.height / 2).
		PaddingBottom(m.height / 2)

	return layout.Render(s)
}

func main() {
	p := tea.NewProgram(intiailizeModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Something went wrong.")
		os.Exit(1)
	}
}
