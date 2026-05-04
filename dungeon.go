package main

import "math/rand"

type BspNode struct {
	X, Y, W, H  int
	Left, Right *BspNode
}

type Dungeon struct {
	Width, Height int
	Tiles         [][]Tile
	Rooms         []Room
}

type Room struct {
	X, Y, W, H int
	Center     Position
}

func NewDungeon(w, h int) *Dungeon {
	d := &Dungeon{Width: w, Height: h}
	d.Tiles = make([][]Tile, h)
	for y := 0; y < h; y++ {
		d.Tiles[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			d.Tiles[y][x] = TileWall
		}
	}
	d.generate()
	return d
}

func (d *Dungeon) generate() {
	// BSP: делим область на под-области
	root := &BspNode{X: 1, Y: 1, W: d.Width - 2, H: d.Height - 2}
	d.split(root, 5) // глубина рекурсии 5 — 2^5 = 32 комнаты максимум
	d.createRooms(root)
	d.connectRooms()
	d.placeStairs()
}

func (d *Dungeon) split(node *BspNode, depth int) {
	if depth <= 0 {
		return
	}
	if node.W < 10 || node.H < 10 {
		return
	}

	// Решаем, делить горизонтально или вертикально
	horizontal := false
	if node.W >= node.H && node.W > 12 {
		horizontal = false // вертикальное деление
	} else if node.H >= node.W && node.H > 12 {
		horizontal = true
	} else if rand.Intn(2) == 0 {
		horizontal = true
	}

	maxSplit := 0
	if horizontal {
		maxSplit = node.H - 6
		if maxSplit < 3 {
			return
		}
		split := 3 + rand.Intn(maxSplit-2)
		node.Left = &BspNode{X: node.X, Y: node.Y, W: node.W, H: split}
		node.Right = &BspNode{X: node.X, Y: node.Y + split, W: node.W, H: node.H - split}
	} else {
		maxSplit = node.W - 6
		if maxSplit < 3 {
			return
		}
		split := 3 + rand.Intn(maxSplit-2)
		node.Left = &BspNode{X: node.X, Y: node.Y, W: split, H: node.H}
		node.Right = &BspNode{X: node.X + split, Y: node.Y, W: node.W - split, H: node.H}
	}
	d.split(node.Left, depth-1)
	d.split(node.Right, depth-1)
}

func (d *Dungeon) createRooms(node *BspNode) {
	if node == nil {
		return
	}
	if node.Left == nil && node.Right == nil {
		// Создаём комнату, только если область достаточно велика
		if node.W < 6 || node.H < 6 {
			return
		}
		rw := 4 + rand.Intn(node.W-5)
		rh := 4 + rand.Intn(node.H-5)
		// Дополнительная страховка от краевого случая
		if node.W-rw <= 0 || node.H-rh <= 0 {
			return
		}
		rx := node.X + rand.Intn(node.W-rw)
		ry := node.Y + rand.Intn(node.H-rh)
		room := Room{X: rx, Y: ry, W: rw, H: rh, Center: Position{X: rx + rw/2, Y: ry + rh/2}}
		d.Rooms = append(d.Rooms, room)
		for y := ry; y < ry+rh; y++ {
			for x := rx; x < rx+rw; x++ {
				d.Tiles[y][x] = TileFloor
			}
		}
		return
	}
	d.createRooms(node.Left)
	d.createRooms(node.Right)
}

func (d *Dungeon) connectRooms() {
	// Соединяем каждую комнату со следующей (линейный порядок — просто и работает)
	for i := 1; i < len(d.Rooms); i++ {
		a := d.Rooms[i-1].Center
		b := d.Rooms[i].Center
		d.carveCorridor(a, b)
	}
}

func (d *Dungeon) carveCorridor(a, b Position) {
	x, y := a.X, a.Y
	for x != b.X {
		d.Tiles[y][x] = TileFloor
		if x < b.X {
			x++
		} else {
			x--
		}
	}
	for y != b.Y {
		d.Tiles[y][x] = TileFloor
		if y < b.Y {
			y++
		} else {
			y--
		}
	}
}

func (d *Dungeon) placeStairs() {
	if len(d.Rooms) == 0 {
		return
	}
	lastRoom := d.Rooms[len(d.Rooms)-1]
	d.Tiles[lastRoom.Center.Y][lastRoom.Center.X] = TileStairsDown
}
