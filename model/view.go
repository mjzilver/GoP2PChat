package model

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) View() string {
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

	return lipgloss.JoinVertical(lipgloss.Top, titleStyle.Render(m.renderTitle()), mainBody, inputStyle.Render(m.renderInput()))
}

func (m *Model) renderTitle() string {
	return title + " - Connected as: " + m.nick
}

func (m *Model) renderMessages() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var msgs string
	for _, msg := range m.messages {
		msgs += msg.String() + "\n"
	}
	return msgs
}

func (m *Model) renderPeers() string {
	if m.Peer.addr == "" {
		return "No peer connected"
	}
	return fmt.Sprintf("Connected peers: \n%s (%s)", m.Peer.Nick, m.Peer.addr)
}

func (m *Model) renderInput() string {
	return lipgloss.NewStyle().Width(m.width).Render("> " + m.input)
}
