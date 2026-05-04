package main

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type GameModel struct {
	Game          *GameView
	Turning       bool // true = ход монстров, false = ход игрока
	ShowInventory bool
	GameOver      bool
	GameWon       bool
	Width, Height int
}

func NewGame() *GameModel {
	d := NewDungeon(40, 20)
	// Ставим игрока в центр первой комнаты
	startPos := Position{X: 1, Y: 1}
	if len(d.Rooms) > 0 {
		startPos = d.Rooms[0].Center
	}
	player := NewPlayer(startPos)
	monsters := spawnMonsters(d, 1)
	items := spawnItems(d, 1)

	gv := &GameView{
		Dungeon:  d,
		Player:   player,
		Monsters: monsters,
		Items:    items,
		Messages: []string{"Добро пожаловать в подземелье!", "wasd — движение, пробел — атака, i — инвентарь, q — выход"},
		Floor:    1,
	}
	gv.ComputeFOV()

	return &GameModel{Game: gv, Turning: false}
}

func (m *GameModel) Init() tea.Cmd {
	return nil
}

func (m *GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.GameOver || m.GameWon {
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "q" {
			return m, tea.Quit
		}
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "r" {
			return NewGame(), nil
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "i":
			m.ShowInventory = !m.ShowInventory
			return m, nil
		}

		if m.Turning {
			return m, nil
		} // ждём хода монстров

		// --- Движение ---
		dx, dy := 0, 0
		switch msg.String() {
		case "w", "up":
			dy = -1
		case "s", "down":
			dy = 1
		case "a", "left":
			dx = -1
		case "d", "right":
			dx = 1
		case " ":
			// Атака в направлении последнего движения (упрощённо — атакуем монстра рядом)
			m.attackMonster()
			m.Turning = true
			return m, m.enemyTurn()
		}

		if dx != 0 || dy != 0 {
			m.movePlayer(dx, dy)
			m.Turning = true
			return m, m.enemyTurn()
		}

	case enemyTurnMsg:
		m.handleEnemyTurn()
		return m, nil
	}

	return m, nil
}

func (m *GameModel) movePlayer(dx, dy int) {
	p := m.Game.Player
	nx, ny := p.Position.X+dx, p.Position.Y+dy
	d := m.Game.Dungeon

	// Проверка границ и проходимости
	if nx < 0 || nx >= d.Width || ny < 0 || ny >= d.Height {
		return
	}
	if d.Tiles[ny][nx] == TileWall {
		return
	}

	// Проверка, не стоит ли на этой клетке живой монстр
	for _, monster := range m.Game.Monsters {
		if monster.Alive && monster.Position.X == nx && monster.Position.Y == ny {
			// Столкновение с монстром — атакуем его
			m.attackMonsterAt(nx, ny)
			return
		}
	}

	// Двигаем игрока
	p.Position.X = nx
	p.Position.Y = ny

	// Подбираем предмет
	for i := 0; i < len(m.Game.Items); i++ {
		it := m.Game.Items[i]
		if it.Position.X == nx && it.Position.Y == ny {
			m.pickupItem(i)
			break
		}
	}

	// Проверка лестницы
	if d.Tiles[ny][nx] == TileStairsDown {
		m.nextFloor()
	}

	m.Game.ComputeFOV()
}

func (m *GameModel) attackMonster() {
	// Атакуем первого монстра рядом с игроком
	px, py := m.Game.Player.Position.X, m.Game.Player.Position.Y
	for _, monster := range m.Game.Monsters {
		if !monster.Alive {
			continue
		}
		mx, my := monster.Position.X, monster.Position.Y
		if abs(mx-px)+abs(my-py) == 1 {
			m.dealDamageToMonster(monster)
			return
		}
	}
	m.Game.Messages = append(m.Game.Messages, "Рядом нет врагов!")
}

func (m *GameModel) attackMonsterAt(x, y int) {
	for _, monster := range m.Game.Monsters {
		if monster.Alive && monster.Position.X == x && monster.Position.Y == y {
			m.dealDamageToMonster(monster)
			return
		}
	}
}

func (m *GameModel) dealDamageToMonster(monster *Monster) {
	dmg := m.Game.Player.AttackDamage()
	monster.HP -= dmg
	msg := fmt.Sprintf("Вы атаковали %s! Нанесено %d урона. HP: %d/%d", monster.Name, dmg, monster.HP, monster.MaxHP)
	if monster.HP <= 0 {
		monster.Alive = false
		msg = fmt.Sprintf("Вы убили %s! +%d опыта", monster.Name, monster.MaxHP)
		m.Game.Player.AddExp(monster.MaxHP)
	}
	m.Game.Messages = append(m.Game.Messages, msg)
}

func (m *GameModel) pickupItem(idx int) {
	item := m.Game.Items[idx]
	p := m.Game.Player

	if item.Heal > 0 {
		// Зелье — сразу используем
		healed := item.Heal
		p.HP += healed
		if p.HP > p.MaxHP {
			p.HP = p.MaxHP
		}
		m.Game.Messages = append(m.Game.Messages, fmt.Sprintf("Вы выпили %s! +%d HP", item.Name, healed))
	} else if item.Damage > 0 {
		// Оружие
		if p.Weapon != nil {
			m.Game.Items = append(m.Game.Items, p.Weapon) // выбросить старое
			m.Game.Items[len(m.Game.Items)-1].Position = p.Position
		}
		p.Weapon = item
		m.Game.Messages = append(m.Game.Messages, fmt.Sprintf("Вы взяли %s! Урон +%d", item.Name, item.Damage))
	} else if item.Defense > 0 {
		if p.Armor != nil {
			m.Game.Items = append(m.Game.Items, p.Armor)
			m.Game.Items[len(m.Game.Items)-1].Position = p.Position
		}
		p.Armor = item
		m.Game.Messages = append(m.Game.Messages, fmt.Sprintf("Вы надели %s! Защита +%d", item.Name, item.Defense))
	}

	// Удаляем предмет с карты
	m.Game.Items = append(m.Game.Items[:idx], m.Game.Items[idx+1:]...)
}

func (m *GameModel) nextFloor() {
	m.Game.Floor++
	d := NewDungeon(40, 20)
	m.Game.Dungeon = d
	if len(d.Rooms) > 0 {
		m.Game.Player.Position = d.Rooms[0].Center
	}
	m.Game.Monsters = spawnMonsters(d, m.Game.Floor)
	m.Game.Items = spawnItems(d, m.Game.Floor)
	m.Game.Messages = append(m.Game.Messages, fmt.Sprintf("Вы спустились на этаж %d! Будьте осторожны...", m.Game.Floor))
	m.Game.ComputeFOV()
}

func (m *GameModel) enemyTurn() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return enemyTurnMsg{}
	})
}

type enemyTurnMsg struct{}

func (m *GameModel) handleEnemyTurn() {
	p := m.Game.Player
	for _, monster := range m.Game.Monsters {
		if !monster.Alive {
			continue
		}

		// Проверка расстояния до игрока
		dx := p.Position.X - monster.Position.X
		dy := p.Position.Y - monster.Position.Y
		dist := abs(dx) + abs(dy)

		if dist == 1 {
			// Атака
			monsterDmg := monster.Damage + rand.Intn(3) // +0..2 рандом
			p.TakeDamage(monsterDmg)
			m.Game.Messages = append(m.Game.Messages, fmt.Sprintf("%s атакует вас! Нанесено %d урона (ваше HP: %d/%d)", monster.Name, monsterDmg, p.HP, p.MaxHP))
			if !p.IsAlive() {
				m.GameOver = true
				m.Game.Messages = append(m.Game.Messages, "ВЫ ПОГИБЛИ В ПОДЗЕМЕЛЬЕ! Нажмите R для рестарта или Q для выхода.")
				return
			}
		} else if dist <= 6 {
			// Движение к игроку (BFS — упрощённо: шаг в сторону игрока)
			nx, ny := monster.Position.X, monster.Position.Y
			if abs(dx) >= abs(dy) {
				if dx > 0 {
					nx++
				} else {
					nx--
				}
			} else {
				if dy > 0 {
					ny++
				} else {
					ny--
				}
			}
			// Проверяем, что клетка свободна
			walkable := true
			if nx < 0 || nx >= m.Game.Dungeon.Width || ny < 0 || ny >= m.Game.Dungeon.Height {
				walkable = false
			} else if m.Game.Dungeon.Tiles[ny][nx] == TileWall {
				walkable = false
			} else if nx == p.Position.X && ny == p.Position.Y {
				walkable = false // не ходим на игрока (атакуем, если рядом)
			}
			for _, other := range m.Game.Monsters {
				if other != monster && other.Alive && other.Position.X == nx && other.Position.Y == ny {
					walkable = false
				}
			}
			if walkable {
				monster.Position.X = nx
				monster.Position.Y = ny
			}
		}
		// Если далеко — монстр стоит на месте
	}
	m.Turning = false
	m.Game.ComputeFOV()
}

func (m *GameModel) View() string {
	if m.GameOver {
		return m.Game.Render() + "\n" + styleInfo.Render("GAME OVER. R — рестарт, Q — выход")
	}
	if m.ShowInventory {
		return m.renderInventory()
	}
	return m.Game.Render()
}

func (m *GameModel) renderInventory() string {
	p := m.Game.Player
	inv := "=== ИНВЕНТАРЬ ===\n"
	if p.Weapon != nil {
		inv += fmt.Sprintf("Оружие: %s (урон +%d)\n", p.Weapon.Name, p.Weapon.Damage)
	} else {
		inv += "Оружие: кулаки (урон 2)\n"
	}
	if p.Armor != nil {
		inv += fmt.Sprintf("Броня: %s (защита +%d)\n", p.Armor.Name, p.Armor.Defense)
	} else {
		inv += "Броня: нет\n"
	}
	inv += fmt.Sprintf("HP: %d/%d\n", p.HP, p.MaxHP)
	inv += fmt.Sprintf("Уровень: %d\n", p.Level)
	inv += fmt.Sprintf("Опыт: %d/%d\n", p.Experience, p.ExpToNext())
	inv += fmt.Sprintf("Урон: %d | Защита: %d\n", p.AttackDamage(), p.DefenseRating())
	inv += "\nНажмите I чтобы закрыть"
	return inv
}
