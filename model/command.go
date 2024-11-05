package model

import (
	"fmt"
	"strings"
	"time"
)

func (m *Model) handleCommand() error {
	switch {
	case strings.HasPrefix(m.input, "/nick "):
		m.nick = strings.TrimPrefix(m.input, "/nick ")
		return nil

	case strings.HasPrefix(m.input, "/connect "):
		return m.connect(strings.TrimPrefix(m.input, "/connect "))

	case m.input == "/help":
		m.sendHelp()
		return nil

	default:
		return m.sendMessage()
	}
}

func (m *Model) connect(peerAddr string) error {
	m.Peer.addr = peerAddr
	m.sendNickRequest()
	m.SendNickResponse()
	return nil
}

func (m *Model) sendMessage() error {
	if m.nick == defaultNick {
		return fmt.Errorf("please set a nickname first. Use /nick <nickname>")
	}

	newMsg := Message{
		Text:      m.input,
		Timestamp: time.Now(),
		Nick:      m.nick,
		Type:      ChatMessage,
	}

	go sendMessage(m, newMsg)

	m.mu.Lock()
	m.messages = append(m.messages, newMsg)
	m.mu.Unlock()
	return nil
}

func (m *Model) sendNickRequest() {
	if m.Peer.addr == "" {
		m.SendError("No peer address set. Use /connect <peer-address>")
		return
	}

	msg := Message{
		Type:      NickRequest,
		Timestamp: time.Now(),
		Nick:      m.nick,
	}

	go sendMessage(m, msg)
}

func (m *Model) SendNickResponse() {
	if m.Peer.addr == "" {
		m.SendError("No peer address set. Use /connect <peer-address>")
		return
	}

	msg := Message{
		Type:      NickResponse,
		Timestamp: time.Now(),
		Nick:      m.nick,
	}

	go sendMessage(m, msg)
}
