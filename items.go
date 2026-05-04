package main

import "math/rand"

var possibleItems = []Item{
	{Name: "Короткий меч", Rune: '/', Damage: 3},
	{Name: "Длинный меч", Rune: '/', Damage: 5},
	{Name: "Секира", Rune: '/', Damage: 8},
	{Name: "Кожаная броня", Rune: '[', Defense: 2},
	{Name: "Кольчуга", Rune: '[', Defense: 4},
	{Name: "Латы", Rune: '[', Defense: 7},
	{Name: "Малое зелье лечения", Rune: '!', Heal: 5},
	{Name: "Большое зелье лечения", Rune: '!', Heal: 15},
}

func spawnItems(d *Dungeon, level int) []*Item {
	count := 2 + level
	items := make([]*Item, 0, count)
	for i := 0; i < count; i++ {
		if len(d.Rooms) < 2 {
			break
		}
		roomIdx := rand.Intn(len(d.Rooms))
		room := d.Rooms[roomIdx]
		x := room.X + 1 + rand.Intn(room.W-2)
		y := room.Y + 1 + rand.Intn(room.H-2)

		item := possibleItems[rand.Intn(len(possibleItems))]
		item.Position = Position{X: x, Y: y}
		items = append(items, &item)
	}
	return items
}
