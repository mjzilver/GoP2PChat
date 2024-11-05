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

	// Create the Bubble Tea program
	p := tea.NewProgram(&m)

	// Start UDP server in a goroutine, passing the program reference
	go startUDPServer(localAddr, p)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}

func startUDPServer(localAddr string, p *tea.Program) {
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

		// Send the message to the Bubble Tea program
		p.Send(NewMessageMsg{msg: msg})
	}
}

type NewMessageMsg struct {
	msg message
}

func (m *model) Init() tea.Cmd {
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
				// if starts with / send to commands function

				if strings.HasPrefix(m.input, "/nick ") {
					m.nick = strings.TrimPrefix(m.input, "/nick ")
					m.input = ""
					return m, nil
				}

				if strings.HasPrefix(m.input, "/connect") {
					// connect to a different peer
					// example: /connect 192.168.1.1:3000
					peerAddr := strings.TrimPrefix(m.input, "/connect ")
					m.peer.addr = peerAddr

					m.input = ""
					return m, nil
				}

				if m.nick == defaultNick {
					fmt.Println("Please set a nickname first. Use /nick <nickname>")
					m.input = ""
					return m, nil
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

				m.input = ""
			}
		default:
			m.input += msg.String()
		}
	case NewMessageMsg:
		m.mu.Lock()
		m.messages = append(m.messages, msg.msg)
		m.mu.Unlock()
		return m, nil
	}
	return m, nil
}

func (m *model) View() string {
	var b strings.Builder
	b.WriteString("P2P Chat\n\n")

	m.mu.Lock()
	for _, msg := range m.messages {
		b.WriteString(msg.String())
		b.WriteString("\n")
	}
	m.mu.Unlock()

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

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshaling message:", err)
		return
	}
	_, err = conn.Write(msgBytes)
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}
