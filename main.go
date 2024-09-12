package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"slices"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	mplusFaceSource *text.GoTextFaceSource
	debugColor	  color.Color
	colorWhite	  color.Color
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s
	debugColor = color.RGBA{0xff, 0x80, 0x0D, 0}
	colorWhite = color.RGBA{0xfa, 0xfa, 0xfa, 0xff}
}

const (
	screenWidth  = 640
	screenHeight = 480
	tileSize = 40
	debug = false
	
)

type Board struct {
	W, H int
	b    []int
	m    []bool
}

func (b *Board) At(x int, y int) (int, bool) {
	return b.b[y*b.W+x], b.m[y*b.W+x]
}

func (b *Board) Mark(x int, y int, v bool) {
	b.m[y*b.W+x] = v
}

func drawRect(screen *ebiten.Image, ax float32, ay float32, bx float32, by float32, thickness float32, clr color.Color) {
	antialias := false
	vector.StrokeLine(screen, ax, ay, ax, by, thickness, clr, antialias)
	vector.StrokeLine(screen, ax, by, bx, by, thickness, clr, antialias)
	vector.StrokeLine(screen, bx, by, bx, ay, thickness, clr, antialias)
	vector.StrokeLine(screen, bx, ay, ax, ay, thickness, clr, antialias)		
}

func (b *Board) centerPos(x, y int) (float32, float32) {
	ulx := (screenWidth - b.W*tileSize) / 2
	uly := (screenHeight - b.H*tileSize) / 2

	cx := float32(ulx+x*tileSize) + 0.5 * tileSize
	cy := float32(uly+y*tileSize) + 0.5 * tileSize
	return cx, cy
}

func (b *Board) Draw(screen *ebiten.Image) {
	ulx := (screenWidth - b.W*tileSize) / 2
	uly := (screenHeight - b.H*tileSize) / 2

	for x := 0; x < b.W; x++ {
		for y := 0; y < b.H; y++ {
			
			n, m := b.At(x, y)
			cx, cy := float32(ulx+x*tileSize), float32(uly+y*tileSize)

			clr := colorWhite
			if m {
				clr = 	color.RGBA{0x2a, 0x2a, 0x2a, 0xff}
			}
			
			op := &text.DrawOptions{}
			op.GeoM.Translate(
				float64(cx + 0.2 * tileSize), 
				float64(cy - 0.23 * tileSize))
			op.ColorScale.ScaleWithColor(clr)

			
			text.Draw(screen, strconv.Itoa(n), &text.GoTextFace{
				Source: mplusFaceSource,
				Size:   tileSize,
			}, op)

			if debug {
				drawRect(screen, cx, cy, cx+tileSize, cy+tileSize, 1, debugColor)
				px, py := b.centerPos(x, y)
				drawRect(screen, px-1, py-1, px+1, py+1, 1, debugColor)
			}
		}
	}
}

func (b *Board) ClickPos(x, y int) (bool, int, int) {
	cx := (x - (screenWidth-b.W*tileSize)/2) / tileSize
	cy := (y - (screenHeight-b.H*tileSize)/2) / tileSize

	if cx >= 0 && cx < b.W && cy >= 0 && cy < b.H {
		return true, cx, cy
	}
	return false, 0, 0
}

type Game struct {
	hist []Board
	mx, my float32 // mouse position
}

func (g *Game) b() *Board {
	return &g.hist[len(g.hist)-1]
}

func (g *Game) doSel(cx, cy float32) {
	lx, rx, ly, ry := min(g.mx, cx), max(g.mx, cx), min(g.my, cy), max(g.my, cy)

	b := g.b()
	total := 0
	sx, sy := []int{}, []int{}

	for x := 0; x < b.W; x++ {
		for y := 0; y < b.H; y++ {
			n, m := b.At(x, y)
			if m {
				continue
			}
			nx, ny := b.centerPos(x, y)
			if nx >= lx && nx <= rx && ny >= ly && ny <= ry {
				total += n
				sx, sy = append(sx, x), append(sy, y)
			}
		}
	}

	if total == 0 || total % 10 != 0 {
		return
	}

	nb := *b
	nb.b = slices.Clone(b.b)
	nb.m = slices.Clone(b.m)
	for i := 0; i < len(sx); i++ {
		nb.Mark(sx[i], sy[i], true)
	}

	g.hist = append(g.hist, nb)
}

func (g *Game) Update() error {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		icx, icy := ebiten.CursorPosition()
		cx, cy := float32(icx), float32(icy)
		g.doSel(cx, cy)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		g.mx, g.my = float32(cx), float32(cy)	
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && len(g.hist) > 1 {
		g.hist = g.hist[:len(g.hist)-1]
	}
		
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.b().Draw(screen)

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		cx, cy := ebiten.CursorPosition()
		bx, by := float32(cx), float32(cy)
		drawRect(screen, g.mx, g.my, bx, by, 1, colorWhite)		
	}
	ebitenutil.DebugPrint(screen, "Press  ARROW LEFT to UNDO")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello there")

	board := Board{
		W: 5, 
		H: 5, 
		b: []int{4, 6, 7, 8, 1, 3, 8, 4, 1, 9, 5, 7, 2, 1, 8, 3, 5, 5, 9, 2, 9, 5, 8, 2, 8},
		m: make([]bool, 5*5),
	}
	g := Game{
		hist: []Board{board},
	}

	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
