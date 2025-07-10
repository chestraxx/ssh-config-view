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
	Host         string
	HostName     string
	User         string
	Port         string
	IdentityFile string
}

// item implements list.Item for Bubbletea

type item SSHHost

func (i item) Title() string       { return i.Host }
func (i item) Description() string { return i.HostName }
func (i item) FilterValue() string { return i.Host }

// model for Bubbletea

type mode int

const (
	modeNormal mode = iota
	modeDetails
	modeEdit
	modeConfirm
	modeDeleteConfirm
	modeDeleted
)

type model struct {
	list     list.Model
	selected *SSHHost
	mode     mode

	// Edit mode state
	editHost  SSHHost
	editField int // 0=Host, 1=HostName, 2=User, 3=Port

	// Track selected index in the list
	selectedIndex int
	confirmMsg    string // message to show in confirmation mode
	deleteIndex   int    // index to delete
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
	return model{list: l, mode: modeNormal}
}

func (m model) Init() tea.Cmd {
	return nil
}

func saveHostsToFile(hosts []SSHHost, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, h := range hosts {
		fmt.Fprintf(f, "Host %s\n", h.Host)
		if h.HostName != "" {
			fmt.Fprintf(f, "  HostName %s\n", h.HostName)
		}
		if h.User != "" {
			fmt.Fprintf(f, "  User %s\n", h.User)
		}
		if h.Port != "" {
			fmt.Fprintf(f, "  Port %s\n", h.Port)
		}
		if h.IdentityFile != "" {
			fmt.Fprintf(f, "  IdentityFile %s\n", h.IdentityFile)
		}
		fmt.Fprintln(f)
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case modeEdit:
			switch msg.String() {
			case "tab", "down":
				m.editField = (m.editField + 1) % 4
				return m, nil
			case "shift+tab", "up":
				m.editField = (m.editField + 3) % 4
				return m, nil
			case "esc":
				m.mode = modeDetails
				return m, nil
			case "enter":
				// Save changes to list and file
				m.list.SetItem(m.selectedIndex, item(m.editHost))
				m.selected = &m.editHost // update selected pointer
				// Save to file
				items := m.list.Items()
				hosts := make([]SSHHost, len(items))
				for i, it := range items {
					hosts[i] = SSHHost(it.(item))
				}
				usr, err := user.Current()
				if err == nil {
					sshConfigPath := filepath.Join(usr.HomeDir, ".ssh", "config")
					err = saveHostsToFile(hosts, sshConfigPath)
					if err != nil {
						m.confirmMsg = "Error saving SSH config! Press any key."
					} else {
						m.confirmMsg = "Changes saved! Press any key to continue."
					}
				} else {
					m.confirmMsg = "Error saving SSH config! Press any key."
				}
				m.mode = modeConfirm
				return m, nil
			}
			// Handle text input
			if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
				ch := msg.String()
				switch m.editField {
				case 0:
					m.editHost.Host += ch
				case 1:
					m.editHost.HostName += ch
				case 2:
					m.editHost.User += ch
				case 3:
					m.editHost.Port += ch
				}
				return m, nil
			}
			// Handle backspace
			if msg.Type == tea.KeyBackspace || msg.Type == tea.KeyDelete {
				switch m.editField {
				case 0:
					if len(m.editHost.Host) > 0 {
						m.editHost.Host = m.editHost.Host[:len(m.editHost.Host)-1]
					}
				case 1:
					if len(m.editHost.HostName) > 0 {
						m.editHost.HostName = m.editHost.HostName[:len(m.editHost.HostName)-1]
					}
				case 2:
					if len(m.editHost.User) > 0 {
						m.editHost.User = m.editHost.User[:len(m.editHost.User)-1]
					}
				case 3:
					if len(m.editHost.Port) > 0 {
						m.editHost.Port = m.editHost.Port[:len(m.editHost.Port)-1]
					}
				}
				return m, nil
			}
			return m, nil
		case modeDetails:
			// Allow edit with 'e'
			if msg.String() == "e" && m.selected != nil {
				m.mode = modeEdit
				m.editHost = *m.selected
				m.editField = 0
				return m, nil
			}
			// Allow delete with 'd'
			if msg.String() == "d" && m.selected != nil {
				m.mode = modeDeleteConfirm
				m.confirmMsg = "Delete this host? (y/n)"
				m.deleteIndex = m.selectedIndex
				return m, nil
			}
			// Return to list on any key
			m.selected = nil
			m.mode = modeNormal
			return m, nil
		case modeNormal:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				if sel, ok := m.list.SelectedItem().(item); ok {
					h := SSHHost(sel)
					m.selected = &h
					m.selectedIndex = m.list.Index()
					m.mode = modeDetails
				}
			}
		case modeConfirm:
			// Any key returns to details mode
			m.mode = modeDetails
			return m, nil
		case modeDeleteConfirm:
			if msg.String() == "y" {
				// Delete the item
				items := m.list.Items()
				if m.deleteIndex >= 0 && m.deleteIndex < len(items) {
					newItems := append(items[:m.deleteIndex], items[m.deleteIndex+1:]...)
					m.list.SetItems(newItems)
					// Save to file
					hosts := make([]SSHHost, len(newItems))
					for i, it := range newItems {
						hosts[i] = SSHHost(it.(item))
					}
					usr, err := user.Current()
					if err == nil {
						sshConfigPath := filepath.Join(usr.HomeDir, ".ssh", "config")
						err = saveHostsToFile(hosts, sshConfigPath)
						if err != nil {
							m.confirmMsg = "Error saving after delete! Press any key."
						} else {
							m.confirmMsg = "Host deleted! Press any key."
						}
					} else {
						m.confirmMsg = "Error saving after delete! Press any key."
					}
				}
				m.mode = modeDeleted
				return m, nil
			} else if msg.String() == "n" {
				m.mode = modeDetails
				return m, nil
			}
			return m, nil
		case modeDeleted:
			// Any key returns to list view
			m.selected = nil
			m.mode = modeNormal
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.mode {
	case modeEdit:
		return m.editView()
	case modeDetails:
		if m.selected != nil {
			s := m.selected
			popup := lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).Width(40).Render(
				fmt.Sprintf(
					"Host: %s\nHostName: %s\nUser: %s\nPort: %s\n\nPress 'e' to edit.\nPress 'd' to delete.\nAny key to return.",
					s.Host, s.HostName, s.User, s.Port,
				),
			)
			return popup
		}
		return ""
	case modeConfirm:
		return lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).Width(40).Render(m.confirmMsg)
	case modeDeleteConfirm:
		return lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).Width(40).Render(m.confirmMsg)
	case modeDeleted:
		return lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).Width(40).Render(m.confirmMsg)
	default:
		return m.list.View() + "\nPress q to quit, Enter for details."
	}
}

func (m model) editView() string {
	fields := []struct {
		label string
		value string
	}{
		{"Host", m.editHost.Host},
		{"HostName", m.editHost.HostName},
		{"User", m.editHost.User},
		{"Port", m.editHost.Port},
	}
	lines := make([]string, len(fields))
	for i, f := range fields {
		cursor := " "
		if m.editField == i {
			cursor = ">"
		}
		lines[i] = fmt.Sprintf("%s %s: %s", cursor, f.label, f.value)
	}
	return lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).Width(50).Render(
		"Edit SSH Host (\nTab/Shift+Tab to move,\nEnter to save,\nEsc to cancel):\n\n" +
			strings.Join(lines, "\n"),
	)
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

	// Check write permission
	f, err := os.OpenFile(sshConfigPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("Cannot write to %s: %v\nPlease check file permissions.\n", sshConfigPath, err)
		os.Exit(1)
	}
	f.Close()

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
			} else if strings.HasPrefix(line, "IdentityFile ") {
				currentHost.IdentityFile = strings.TrimSpace(strings.TrimPrefix(line, "IdentityFile "))
			}

			if currentHost.Port == "" {
				currentHost.Port = "22"
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
