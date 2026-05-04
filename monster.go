package main

import "math/rand"

var monsterTemplates = []Monster{
	{Name: "Крыса", HP: 4, MaxHP: 4, Damage: 1, Defense: 0, Rune: 'r', Alive: true},
	{Name: "Скелет", HP: 8, MaxHP: 8, Damage: 3, Defense: 1, Rune: 's', Alive: true},
	{Name: "Орк", HP: 14, MaxHP: 14, Damage: 5, Defense: 2, Rune: 'O', Alive: true},
	{Name: "Демон", HP: 25, MaxHP: 25, Damage: 8, Defense: 3, Rune: 'D', Alive: true},
}

func spawnMonsters(d *Dungeon, level int) []*Monster {
	count := 3 + level*2 // на первом уровне 5 монстров, на 10-м — 23
	monsters := make([]*Monster, 0, count)

	for i := 0; i < count; i++ {
		// Выбираем комнату (не первую — там спавн игрока)
		if len(d.Rooms) < 2 {
			break
		}
		roomIdx := 1 + rand.Intn(len(d.Rooms)-1)
		room := d.Rooms[roomIdx]

		// Случайная позиция в комнате (не на стене)
		x := room.X + 1 + rand.Intn(room.W-2)
		y := room.Y + 1 + rand.Intn(room.H-2)

		// Выбираем монстра по уровню
		template := monsterTemplates[0]
		if level >= 3 {
			template = monsterTemplates[rand.Intn(2)+1]
		} // скелет или орк
		if level >= 5 {
			template = monsterTemplates[rand.Intn(3)+1]
		} // орк+ или демон
		if level >= 7 {
			template = monsterTemplates[rand.Intn(4)]
		}
		if level >= 10 {
			template = monsterTemplates[3]
		} // только демоны

		m := template
		m.Position = Position{X: x, Y: y}
		// Масштабируем HP и урон с уровнем
		m.HP += level
		m.MaxHP = m.HP
		m.Damage += level / 2
		m.Alive = true
		monsters = append(monsters, &m)
	}
	return monsters
}
