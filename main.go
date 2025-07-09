package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SSHHost struct

type SSHHost struct {
	Host     string
	HostName string
	User     string
	Port     string
}

// item implements list.Item for Bubbletea

type item SSHHost

func (i item) Title() string       { return i.Host }
func (i item) Description() string { return i.HostName }
func (i item) FilterValue() string { return i.Host }

// model for Bubbletea

type model struct {
	list     list.Model
	selected *SSHHost
}

var (
	listHeight = 14
	listWidth  = 60
)

func initialModel(hosts []SSHHost) model {
	items := make([]list.Item, len(hosts))
	for i, h := range hosts {
		items[i] = item(h)
	}
	l := list.New(items, list.NewDefaultDelegate(), listWidth, listHeight)
	l.Title = "SSH Config Hosts"
	return model{list: l}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.selected != nil {
			m.selected = nil
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if sel, ok := m.list.SelectedItem().(item); ok {
				h := SSHHost(sel)
				m.selected = &h
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.selected != nil {
		// Show details popup
		s := m.selected
		popup := lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).Width(40).Render(
			fmt.Sprintf(
				"Host: %s\nHostName: %s\nUser: %s\nPort: %s\n\nPress any key to return.",
				s.Host, s.HostName, s.User, s.Port,
			),
		)
		return popup
	}
	return m.list.View() + "\nPress q to quit, Enter for details."
}

func main() {
	search := flag.String("search", "", "Filter hosts by name")
	flag.StringVar(search, "s", "", "Filter hosts by name (shorthand)")
	flag.Parse()

	// Show loading screen for at least 2 seconds, then proceed
	p := tea.NewProgram(LoadingModel{})
	if _, err := p.StartReturningModel(); err != nil {
		fmt.Println("Error running loading screen:", err)
		os.Exit(1)
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		os.Exit(1)
	}
	sshConfigPath := filepath.Join(usr.HomeDir, ".ssh", "config")

	file, err := os.Open(sshConfigPath)
	if err != nil {
		fmt.Printf("Could not open SSH config file at %s: %v\n", sshConfigPath, err)
		os.Exit(1)
	}
	defer file.Close()

	hosts := []SSHHost{}
	var currentHost *SSHHost
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "Host ") {
			if currentHost != nil {
				hosts = append(hosts, *currentHost)
			}
			hostLine := strings.TrimPrefix(line, "Host ")
			hostLine = strings.TrimSpace(hostLine)
			currentHost = &SSHHost{Host: hostLine}
			continue
		}
		if currentHost != nil {
			if strings.HasPrefix(line, "HostName ") {
				currentHost.HostName = strings.TrimSpace(strings.TrimPrefix(line, "HostName "))
			} else if strings.HasPrefix(line, "User ") {
				currentHost.User = strings.TrimSpace(strings.TrimPrefix(line, "User "))
			} else if strings.HasPrefix(line, "Port ") {
				currentHost.Port = strings.TrimSpace(strings.TrimPrefix(line, "Port "))
			}
		}
	}
	if currentHost != nil {
		hosts = append(hosts, *currentHost)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading SSH config:", err)
		os.Exit(1)
	}

	// Filter hosts if search is provided
	if *search != "" {
		filtered := make([]SSHHost, 0, len(hosts))
		for _, h := range hosts {
			if strings.Contains(h.Host, *search) || strings.Contains(h.HostName, *search) {
				filtered = append(filtered, h)
			}
		}
		hosts = filtered
	}

	p2 := tea.NewProgram(initialModel(hosts))
	if err := p2.Start(); err != nil {
		fmt.Println("Error running Bubbletea program:", err)
		os.Exit(1)
	}
}
