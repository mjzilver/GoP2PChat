package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"p2p-chat/model"
	"p2p-chat/network"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <listening-port>")
		return
	}

	localAddr := os.Args[1]

	m := model.NewModel()
	p := tea.NewProgram(&m)

	go network.StartUDPServer(localAddr, p, &m)

	if _, err := p.Run(); err != nil {
		m.SendError(fmt.Sprintf("Error running program: %v", err))
	}

	m.SendError("Program exited")
}
