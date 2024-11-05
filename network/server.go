package network

import (
	"encoding/json"
	"fmt"
	"net"
	"p2p-chat/model"

	tea "github.com/charmbracelet/bubbletea"
)

func StartUDPServer(localAddr string, p *tea.Program, m *model.Model) {
	addr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		m.SendError(fmt.Sprintf("Failed to resolve address %s: %v", localAddr, err))
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		m.SendError(fmt.Sprintf("Failed to listen on %s: %v", localAddr, err))
		return
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			m.SendError(fmt.Sprintf("Error reading from UDP: %v", err))
			continue
		}

		var msg model.Message
		err = json.Unmarshal(buf[:n], &msg)
		if err != nil {
			m.SendError(fmt.Sprintf("Error unmarshaling message: %v", err))
			continue
		}

		switch msg.Type {

		case model.NickRequest:
			m.Peer.Nick = msg.Nick
			m.SendNickResponse()
		case model.NickResponse:
			m.Peer.Nick = msg.Nick
		case model.ErrorMessage, model.SystemMessage, model.ChatMessage:
			p.Send(model.NewMessageMsg{Msg: msg})
		}
	}
}
