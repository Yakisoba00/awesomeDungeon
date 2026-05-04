package main

type Player struct {
	Position   Position
	HP         int
	MaxHP      int
	Damage     int
	Defense    int
	Experience int
	Level      int
	Inventory  []Item
	Weapon     *Item
	Armor      *Item
}

func NewPlayer(start Position) *Player {
	return &Player{
		Position: start,
		HP:       20,
		MaxHP:    20,
		Damage:   2,
		Defense:  0,
		Level:    1,
	}
}

func (p *Player) AttackDamage() int {
	dmg := p.Damage
	if p.Weapon != nil {
		dmg += p.Weapon.Damage
	}
	return dmg
}

func (p *Player) DefenseRating() int {
	def := p.Defense
	if p.Armor != nil {
		def += p.Armor.Defense
	}
	return def
}

func (p *Player) TakeDamage(amount int) {
	actual := amount - p.DefenseRating()
	if actual < 0 {
		actual = 0
	}
	p.HP -= actual
	if p.HP < 0 {
		p.HP = 0
	}
}

func (p *Player) AddExp(amount int) {
	p.Experience += amount
	for p.Experience >= p.ExpToNext() {
		p.Experience -= p.ExpToNext()
		p.LevelUp()
	}
}

func (p *Player) ExpToNext() int {
	return p.Level * 10 // 10, 20, 30...
}

func (p *Player) LevelUp() {
	p.Level++
	p.MaxHP += 5
	p.HP = p.MaxHP
	p.Damage += 1
	p.Defense++
}

func (p *Player) IsAlive() bool {
	return p.HP > 0
}
