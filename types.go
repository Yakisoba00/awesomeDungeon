package main

type Position struct {
	X, Y int
}

type Tile int

const (
	TileWall       Tile = iota // стена (непроходима)
	TileFloor                  // пол (проходим)
	TileDoor                   // дверь (проходима)
	TileStairsDown             // лестница вниз
)

type Item struct {
	Name     string
	Rune     rune // символ на карте
	Damage   int  // для оружия
	Heal     int  // для зелий
	Defense  int  // для брони
	Position Position
}

type MonsterType int

const (
	MonsterRat MonsterType = iota
	MonsterSkeleton
	MonsterOrc
	MonsterDemon
)

type Monster struct {
	Type     MonsterType
	Name     string
	HP       int
	MaxHP    int
	Damage   int
	Defense  int
	Position Position
	Rune     rune // символ на карте
	Alive    bool
}
