package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleWall    = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // серый
	styleFloor   = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // белый
	styleStairs  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // жёлтый
	stylePlayer  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // зелёный
	styleMonster = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // красный
	styleItem    = lipgloss.NewStyle().Foreground(lipgloss.Color("5")) // пурпурный
	styleHP      = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleInfo    = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // голубой
)

type GameView struct {
	Dungeon  *Dungeon
	Player   *Player
	Monsters []*Monster
	Items    []*Item
	Messages []string // лог сообщений (последние 5)
	Floor    int
	FOV      [][]bool // туман войны (field of view)
}

func (gv *GameView) Render() string {
	var b strings.Builder

	// --- Карта ---
	for y := 0; y < gv.Dungeon.Height; y++ {
		for x := 0; x < gv.Dungeon.Width; x++ {
			ch := ' '
			style := styleFloor

			if !gv.FOV[y][x] {
				b.WriteString(" ")
				continue
			}

			tile := gv.Dungeon.Tiles[y][x]
			switch tile {
			case TileWall:
				ch, style = '#', styleWall
			case TileFloor:
				ch, style = '.', styleFloor
			case TileStairsDown:
				ch, style = '>', styleStairs
			}

			// Игрок — рисуется поверх всего (кроме стен)
			if gv.Player.Position.X == x && gv.Player.Position.Y == y {
				ch, style = '@', stylePlayer
			}

			// Монстры
			for _, m := range gv.Monsters {
				if m.Alive && m.Position.X == x && m.Position.Y == y {
					ch, style = m.Rune, styleMonster
				}
			}

			// Предметы
			for _, it := range gv.Items {
				if it.Position.X == x && it.Position.Y == y {
					if ch == '.' || ch == '>' {
						ch, style = it.Rune, styleItem
					}
				}
			}

			b.WriteString(style.Render(string(ch)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// --- UI-панель ---
	hpColor := styleHP
	hpPercent := float64(gv.Player.HP) / float64(gv.Player.MaxHP)
	if hpPercent > 0.5 {
		hpColor = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	} // зелёный
	if hpPercent <= 0.3 {
		hpColor = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	} // жёлтый

	b.WriteString(styleInfo.Render(fmt.Sprintf("Этаж: %d", gv.Floor)))
	b.WriteString("  ")
	b.WriteString(hpColor.Render(fmt.Sprintf("HP: %d/%d", gv.Player.HP, gv.Player.MaxHP)))
	b.WriteString(fmt.Sprintf("  Ур: %d  Опыт: %d/%d", gv.Player.Level, gv.Player.Experience, gv.Player.ExpToNext()))
	if gv.Player.Weapon != nil {
		b.WriteString(fmt.Sprintf("  Оружие: %s(+%d)", gv.Player.Weapon.Name, gv.Player.Weapon.Damage))
	}
	if gv.Player.Armor != nil {
		b.WriteString(fmt.Sprintf("  Броня: %s(+%d)", gv.Player.Armor.Name, gv.Player.Armor.Defense))
	}
	b.WriteString("\n\n")

	// --- Сообщения ---
	start := 0
	if len(gv.Messages) > 5 {
		start = len(gv.Messages) - 5
	}
	for _, msg := range gv.Messages[start:] {
		b.WriteString(msg + "\n")
	}

	// --- Инвентарь (по нажатию 'i') ---
	// (Вынесу в Update — будет флаг showInventory)

	return b.String()
}

// Туман войны — простой рейкастинг (только то, что видит игрок)
func (gv *GameView) ComputeFOV() {
	gv.FOV = make([][]bool, gv.Dungeon.Height)
	for y := 0; y < gv.Dungeon.Height; y++ {
		gv.FOV[y] = make([]bool, gv.Dungeon.Width)
	}

	px, py := gv.Player.Position.X, gv.Player.Position.Y
	radius := 8

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			dist := dx*dx + dy*dy
			if dist > radius*radius {
				continue
			}

			// Простой raycasting до этой точки

			steps := max(abs(dx), abs(dy))
			if steps == 0 {
				continue
			}

			blocked := false
			for i := 0; i <= steps; i++ {
				cx := px + dx*i/steps
				cy := py + dy*i/steps
				if cx < 0 || cx >= gv.Dungeon.Width || cy < 0 || cy >= gv.Dungeon.Height {
					blocked = true
					break
				}
				if !blocked {
					gv.FOV[cy][cx] = true
				}
				if gv.Dungeon.Tiles[cy][cx] == TileWall {
					blocked = true
				}
			}
		}
	}
	// Сам игрок
	gv.FOV[py][px] = true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
