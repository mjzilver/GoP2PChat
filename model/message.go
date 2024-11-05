package model

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func sendMessage(m *Model, msg Message) {
	if m.Peer.addr == "" {
		m.SendError("No peer address set. Use /connect <peer-address>")
		return
	}

	conn, err := net.Dial("udp", m.Peer.addr)
	if err != nil {
		m.SendError(fmt.Sprintf("Error connecting to peer: %v", err))
		return
	}
	defer conn.Close()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		m.SendError(fmt.Sprintf("Error marshaling message: %v", err))
		return
	}
	_, err = conn.Write(msgBytes)
	if err != nil {
		m.SendError(fmt.Sprintf("Error sending message: %v", err))
		return
	}
}

func (m *Model) SendError(err string) {
	m.Update(NewMessageMsg{
		Msg: Message{
			Text:      err,
			Timestamp: time.Now(),
			Nick:      "Error",
			Type:      ErrorMessage,
		},
	})
}

func (m *Model) sendHelp() {
	m.Update(NewMessageMsg{
		Msg: Message{
			Text:      helpMsg,
			Timestamp: time.Now(),
			Nick:      "System",
			Type:      SystemMessage,
		},
	})
}

func (m *Model) sendWelcome() {
	m.Update(NewMessageMsg{
		Msg: Message{
			Text:      welcomeMsg,
			Timestamp: time.Now(),
			Nick:      "System",
			Type:      SystemMessage,
		},
	})
}
