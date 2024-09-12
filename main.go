package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	mplusFaceSource *text.GoTextFaceSource
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s
}

const (
	screenWidth  = 640
	screenHeight = 480
)

const (
	tileSize = 30
)

type Board struct {
	W, H int
	b    []int
}

func (b *Board) At(x int, y int) int {
	return b.b[y*b.W+x]
}

func (b *Board) Set(x int, y int, v int) {
	b.b[y*b.W+x] = v
}

func (b *Board) Draw(screen *ebiten.Image) {
	ulx := (screenWidth - b.W*tileSize) / 2
	uly := (screenHeight - b.H*tileSize) / 2

	for x := 0; x < b.W; x++ {
		for y := 0; y < b.H; y++ {
			clr := color.RGBA{0xfa, 0xfa, 0xfa, 0}

			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(ulx+x*tileSize), float64(uly+y*tileSize))
			op.ColorScale.ScaleWithColor(clr)

			text.Draw(screen, strconv.Itoa(b.At(x, y)), &text.GoTextFace{
				Source: mplusFaceSource,
				Size:   tileSize,
			}, op)

			// if b.At(x, y) {
			// 	clr = color.RGBA{0xFA, 0xFA, 0xFA, 0}
			// }
			// vector.DrawFilledRect(screen, float32(ulx+x*tileSize), float32(uly+y*tileSize), tileSize, tileSize, clr, false)
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
}

func (g *Game) b() *Board {
	return &g.hist[len(g.hist)-1]
}

func (g *Game) Update() error {
	// if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
	// 	mx, my := ebiten.CursorPosition()
	// 	if ok, cx, cy := g.b().ClickPos(mx, my); ok {
	// 		g.hist = append(g.hist, g.b().Fill(cx, cy))
	// 	}
	// }
	// if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && len(g.hist) > 1 {
	// 	g.hist = g.hist[:len(g.hist)-1]
	// }
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.b().Draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Steps: %d\n Press  ARROW LEFT to UNDO", len(g.hist)-1))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello there")

	board := Board{W: 5, H: 5, b: []int{4, 6, 7, 8, 1, 3, 8, 4, 1, 9, 5, 7, 2, 1, 8, 3, 5, 5, 9, 2, 9, 5, 8, 2, 8}}
	g := Game{
		hist: []Board{board},
	}

	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
