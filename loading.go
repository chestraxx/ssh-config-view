package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type progressMsg int

type LoadingModel struct {
	progress int // 0-100
}

func (m LoadingModel) Init() tea.Cmd {
	return tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg {
		return progressMsg(1)
	})
}

func (m LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case progressMsg:
		if m.progress < 100 {
			m.progress++
			return m, tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg {
				return progressMsg(1)
			})
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m LoadingModel) View() string {
	barWidth := 40
	filled := int(float64(m.progress) / 100 * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat(" ", barWidth-filled)
	return lipgloss.NewStyle().Padding(2, 4).Render(
		fmt.Sprintf("Loading SSH config... Please wait.\n[%s] %d%%", bar, m.progress),
	)
}
