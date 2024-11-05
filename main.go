package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type peer struct {
	addr string
	nick string
}

type message struct {
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Nick      string    `json:"nick"`
}

func (m message) String() string {
	return fmt.Sprintf("[%s] %s: %s", m.Timestamp.Format("15:04"), m.Nick, m.Text)
}

type model struct {
	nick     string
	input    string
	messages []message
	mu       sync.Mutex
	peer     peer
	width    int
	height   int
}

const (
	title       = "Go P2P UDP Chat"
	defaultNick = "Anonymous"
	welcomeMsg  = "Welcome to the Go P2P UDP Chat!"
	helpMsg     = "Type /nick <nickname> to set your nickname. \n/Connect <peer-address> to connect to a peer. \nPress enter to send a message."
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <listening-port>")
		return
	}

	localAddr := os.Args[1]

	m := model{
		nick: defaultNick,
		peer: peer{
			nick: defaultNick,
		},
	}

	p := tea.NewProgram(&m)

	go startUDPServer(localAddr, p, &m)

	if _, err := p.Run(); err != nil {
		m.sendError(fmt.Sprintf("Error running program: %v", err))
	}

	m.sendError("Program exited")
}

func startUDPServer(localAddr string, p *tea.Program, m *model) {
	addr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		m.sendError(fmt.Sprintf("Failed to resolve address %s: %v", localAddr, err))
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		m.sendError(fmt.Sprintf("Failed to listen on %s: %v", localAddr, err))
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			m.sendError(fmt.Sprintf("Error reading from UDP: %v", err))
			continue
		}

		var msg message
		err = json.Unmarshal(buf[:n], &msg)
		if err != nil {
			m.sendError(fmt.Sprintf("Error unmarshaling message: %v", err))
			continue
		}

		p.Send(NewMessageMsg{msg: msg})
	}
}

type NewMessageMsg struct {
	msg message
}

func (m *model) Init() tea.Cmd {
	m.sendWelcome()
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "enter":
			if m.input != "" {
				if err := m.handleCommand(); err != nil {
					m.sendError(err.Error())
				}
				m.input = ""
			}
		default:
			m.input += msg.String()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case NewMessageMsg:
		m.mu.Lock()
		m.messages = append(m.messages, msg.msg)
		m.mu.Unlock()

		return m, nil
	}
	return m, nil
}

func (m *model) handleCommand() error {
	switch {
	case strings.HasPrefix(m.input, "/nick "):
		m.nick = strings.TrimPrefix(m.input, "/nick ")
		return nil

	case strings.HasPrefix(m.input, "/connect "):
		peerAddr := strings.TrimPrefix(m.input, "/connect ")
		m.peer.addr = peerAddr
		return nil

	case m.input == "/help":
		m.sendHelp()
		return nil

	default:
		if m.nick == defaultNick {
			return fmt.Errorf("please set a nickname first. Use /nick <nickname>")
		}

		newMsg := message{
			Text:      m.input,
			Timestamp: time.Now(),
			Nick:      m.nick,
		}

		go sendMessage(m, newMsg)

		m.mu.Lock()
		m.messages = append(m.messages, newMsg)
		m.mu.Unlock()
		return nil
	}
}

func (m *model) View() string {
	mainBodyHeight := m.height - 2

	leftWidth := (m.width / 5) * 4
	rightWidth := m.width - leftWidth

	chatStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(mainBodyHeight).
		Background(lipgloss.Color("#1d1f21")).
		Foreground(lipgloss.Color("#c5c8c6"))

	peerListStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(mainBodyHeight).
		Padding(0, 1).
		Background(lipgloss.Color("#282a36")).
		Foreground(lipgloss.Color("#50fa7b"))

	inputStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(1).
		Background(lipgloss.Color("#44475a")).
		Foreground(lipgloss.Color("#50fa7b"))

	titleStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(1).
		Background(lipgloss.Color("#44475a")).
		Foreground(lipgloss.Color("#f8f8f2")).
		Bold(true).
		Italic(true)

	mainBody := lipgloss.JoinHorizontal(lipgloss.Top, chatStyle.Render(m.renderMessages()), peerListStyle.Render(m.renderPeers()))

	return lipgloss.JoinVertical(lipgloss.Top, titleStyle.Render(title), mainBody, inputStyle.Render(m.renderInput()))
}

func (m *model) renderMessages() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var msgs string
	for _, msg := range m.messages {
		msgs += msg.String() + "\n"
	}
	return msgs
}

func (m *model) renderPeers() string {
	if m.peer.addr == "" {
		return "No peer connected"
	}
	return fmt.Sprintf("Connected peers: \n%s (%s)", m.peer.nick, m.peer.addr)
}

func (m *model) renderInput() string {
	return lipgloss.NewStyle().Width(m.width).Render("> " + m.input)
}

func (m *model) sendError(err string) {
	m.Update(NewMessageMsg{
		msg: message{
			Text:      err,
			Timestamp: time.Now(),
			Nick:      "Error",
		},
	})
}

func (m *model) sendHelp() {
	m.Update(NewMessageMsg{
		msg: message{
			Text:      helpMsg,
			Timestamp: time.Now(),
			Nick:      "System",
		},
	})
}

func (m *model) sendWelcome() {
	m.Update(NewMessageMsg{
		msg: message{
			Text:      welcomeMsg,
			Timestamp: time.Now(),
			Nick:      "System",
		},
	})
}

func sendMessage(m *model, msg message) {
	if m.peer.addr == "" {
		m.sendError("No peer address set. Use /connect <peer-address>")
		return
	}

	conn, err := net.Dial("udp", m.peer.addr)
	if err != nil {
		m.sendError(fmt.Sprintf("Error connecting to peer: %v", err))
		return
	}
	defer conn.Close()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		m.sendError(fmt.Sprintf("Error marshaling message: %v", err))
		return
	}
	_, err = conn.Write(msgBytes)
	if err != nil {
		m.sendError(fmt.Sprintf("Error sending message: %v", err))
		return
	}
}
