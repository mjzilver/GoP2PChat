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
}

const defaultNick = "Anonymous"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <local-addr> <peer-addr>")
		fmt.Println("Example: go run main.go :3000 :3001")
		return
	}

	localAddr := os.Args[1]
	peerAddr := os.Args[2]

	m := model{
		peer: peer{
			addr: peerAddr,
			nick: defaultNick,
		},
		nick: defaultNick,
	}

	go startUDPServer(localAddr, &m)

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}

func startUDPServer(localAddr string, m *model) {
	addr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		fmt.Printf("Failed to resolve address %s: %v\n", localAddr, err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Failed to listen on %s: %v\n", localAddr, err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		// Unmarshal the message
		var msg message
		err = json.Unmarshal(buf[:n], &msg)
		if err != nil {
			fmt.Println("Error unmarshaling message:", err)
			continue
		}

		// Add the message to the model
		message := message{
			Text:      msg.Text,
			Timestamp: msg.Timestamp,
			Nick:      msg.Nick,
		}

		m.mu.Lock()
		m.messages = append(m.messages, message)
		m.mu.Unlock()

		// Update the terminal
		m.Update(newMessageMsg{})
	}
}

type newMessageMsg struct{}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.input != "" {
				if strings.HasPrefix(m.input, "/nick ") {
					m.nick = strings.TrimPrefix(m.input, "/nick ")
					m.input = ""
					return m, nil
				}

				message := message{
					Text:      m.input,
					Timestamp: time.Now(),
					Nick:      m.nick,
				}

				go sendMessage(m, message)
				m.input = ""
			}
		default:
			m.input += msg.String()
		}

	case newMessageMsg:
		// This message is sent from the UDP server goroutine
		// to update the terminal with a new message
		return m, nil
	}
	return m, nil
}

func (m *model) View() string {
	var b strings.Builder
	b.WriteString("P2P Chat\n\n")

	for _, message := range m.messages {
		b.WriteString(message.String())
		b.WriteString("\n")
	}

	b.WriteString("\n> " + m.input)
	return b.String()
}

func sendMessage(m *model, msg message) {
	conn, err := net.Dial("udp", m.peer.addr)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}
	defer conn.Close()

	message := message{
		Text:      msg.Text,
		Timestamp: msg.Timestamp,
		Nick:      m.nick,
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling message:", err)
		return
	}
	_, err = conn.Write(msgBytes)
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}
