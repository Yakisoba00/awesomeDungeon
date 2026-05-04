package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

func main() {
	// Устанавливаем seed для rand
	// rand.Seed(time.Now().UnixNano()) — не нужно в Go 1.20+
	p := tea.NewProgram(NewGame(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
