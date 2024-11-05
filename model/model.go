package model

import (
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Peer struct {
	addr string
	Nick string
}

type MessageType string

const (
	ChatMessage   MessageType = "chat"
	NickRequest   MessageType = "nick_request"
	NickResponse  MessageType = "nick_response"
	ErrorMessage  MessageType = "error"
	SystemMessage MessageType = "system"
)

type Message struct {
	Text      string      `json:"text"`
	Timestamp time.Time   `json:"timestamp"`
	Nick      string      `json:"nick"`
	Type      MessageType `json:"type"`
}

func (m Message) String() string {
	return fmt.Sprintf("[%s] %s: %s", m.Timestamp.Format("15:04"), m.Nick, m.Text)
}

type Model struct {
	nick     string
	input    string
	messages []Message
	mu       sync.Mutex
	Peer     Peer
	width    int
	height   int
}

func NewModel() Model {
	return Model{
		nick:     defaultNick,
		Peer:     Peer{Nick: defaultNick},
		messages: []Message{},
	}
}

const (
	title       = "Go P2P UDP Chat"
	defaultNick = "Anonymous"
	welcomeMsg  = "Welcome to the Go P2P UDP Chat!"
	helpMsg     = "Type /nick <nickname> to set your nickname. \n/Connect <peer-address> to connect to a peer. \nPress enter to send a message."
)

type NewMessageMsg struct {
	Msg Message
}

func (m *Model) Init() tea.Cmd {
	m.sendWelcome()
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					m.SendError(err.Error())
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
		m.messages = append(m.messages, msg.Msg)
		m.mu.Unlock()

		return m, nil
	}
	return m, nil
}
