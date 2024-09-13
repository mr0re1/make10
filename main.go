package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"slices"
	"strconv"
	"strings"

	"make10/pkg"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	mplusFaceSource *text.GoTextFaceSource
	debugColor      color.Color
	colorWhite      color.Color
	dColors         = map[int]color.Color{}
)

func init() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}
	mplusFaceSource = s
	debugColor = color.RGBA{0xff, 0x80, 0x0D, 0}
	colorWhite = color.RGBA{0xfa, 0xfa, 0xfa, 0xff}
	dColors = map[int]color.Color{
		1: color.RGBA{0x2e, 0x8b, 0x57, 0xff},
		2: color.RGBA{0xff, 0xe4, 0xc4, 0xff},
		3: color.RGBA{0xff, 0x00, 0x00, 0xff},
		4: color.RGBA{0xff, 0xff, 0x00, 0xff},
		5: color.RGBA{0x00, 0xff, 0x00, 0xff},
		6: color.RGBA{0xe9, 0x96, 0x7a, 0xff},
		7: color.RGBA{0x00, 0xbf, 0xff, 0xff},
		8: color.RGBA{0xff, 0x80, 0x0d, 0xff},
		9: color.RGBA{0xff, 0x14, 0x93, 0xff},
	}
}

const (
	screenWidth  = 640
	screenHeight = 480
	tileSize     = 40
	debug        = false
)

type Board struct {
	W, H int
	b    []int
	m    []bool
}

func buildBoard(w int, b []int) Board {
	return Board{
		W: w,
		H: len(b) / w,
		b: b,
		m: make([]bool, len(b)),
	}
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

	cx := float32(ulx+x*tileSize) + 0.5*tileSize
	cy := float32(uly+y*tileSize) + 0.5*tileSize
	return cx, cy
}

func (b *Board) Draw(screen *ebiten.Image) {
	ulx := (screenWidth - b.W*tileSize) / 2
	uly := (screenHeight - b.H*tileSize) / 2

	for x := 0; x < b.W; x++ {
		for y := 0; y < b.H; y++ {

			n, m := b.At(x, y)
			cx, cy := float32(ulx+x*tileSize), float32(uly+y*tileSize)

			op := &text.DrawOptions{}
			op.GeoM.Translate(
				float64(cx+0.2*tileSize),
				float64(cy-0.23*tileSize))
			op.ColorScale.ScaleWithColor(dColors[n])
			if m {
				op.ColorScale.ScaleAlpha(0.2)
			}

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
	level  int
	levels []Board
	hist   []Board

	mouse pkg.PosControl
	_log  []string
}

func (g *Game) b() *Board {
	return &g.hist[len(g.hist)-1]
}

func (g *Game) log(msg string) {
	g._log = append([]string{msg}, g._log...)
	if len(g._log) > 30 {
		g._log = g._log[:30]
	}
}

func (g *Game) doSel(s pkg.Pos, f pkg.Pos) {
	g.log(fmt.Sprintf("doSel %v %v", s, f))
	lx, rx, ly, ry := min(s.X, f.X), max(s.X, f.X), min(s.Y, f.Y), max(s.Y, f.Y)

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

	if total != 10 {
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
	g.mouse.Update()
	if ok, s, f := g.mouse.JustFinishedDragging(); ok {
		g.doSel(s, f)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && len(g.hist) > 1 {
		g.hist = g.hist[:len(g.hist)-1]
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		g.level = (g.level + 1) % len(g.levels)
		g.hist = []Board{g.levels[g.level]}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.level = (g.level - 1 + len(g.levels)) % len(g.levels)
		g.hist = []Board{g.levels[g.level]}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.b().Draw(screen)

	if ok, s, f := g.mouse.IsDragging(); ok {
		drawRect(screen, s.X, s.Y, f.X, f.Y, 1, colorWhite)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf(`Level %d of %d
Controls:
N - next level
P - previous level
ARROW LEFT - UNDO`, g.level+1, len(g.levels)))

	if debug {
		ebitenutil.DebugPrintAt(screen, strings.Join(g._log, "\n"), 20, 220)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello there")

	levels := []Board{
		buildBoard(5, []int{6, 3, 2, 2, 3, 4, 3, 2, 4, 1, 1, 7, 1, 3, 6, 1, 6, 4, 7, 2, 3, 1, 3, 3, 2}),
		buildBoard(5, []int{5, 1, 1, 1, 6, 5, 7, 1, 1, 2, 5, 4, 7, 1, 2, 5, 6, 4, 2, 5, 8, 2, 2, 2, 5}),
		buildBoard(5, []int{8, 2, 2, 1, 9, 3, 1, 2, 5, 5, 3, 4, 5, 4, 4, 4, 5, 1, 1, 1, 7, 3, 2, 1, 7}),
		buildBoard(5, []int{2, 5, 7, 3, 7, 8, 5, 3, 9, 1, 2, 6, 4, 7, 3, 8, 3, 7, 4, 6, 5, 5, 1, 2, 7}),
		buildBoard(5, []int{5, 7, 2, 2, 6, 5, 3, 2, 7, 1, 1, 1, 1, 1, 9, 7, 6, 1, 2, 7, 6, 4, 4, 2, 8}),
		buildBoard(5, []int{1, 7, 1, 1, 2, 4, 4, 5, 1, 8, 6, 2, 2, 2, 4, 2, 5, 1, 2, 2, 9, 1, 4, 6, 8}),
		buildBoard(5, []int{2, 6, 2, 1, 9, 2, 4, 2, 2, 2, 2, 1, 4, 3, 8, 3, 1, 3, 5, 1, 7, 2, 8, 8, 2}),
		buildBoard(5, []int{5, 3, 5, 1, 7, 5, 1, 1, 1, 1, 3, 3, 1, 2, 7, 1, 3, 2, 5, 2, 3, 2, 2, 3, 1}),
		buildBoard(5, []int{2, 8, 9, 1, 4, 7, 1, 1, 1, 1, 3, 7, 3, 7, 1, 1, 3, 1, 5, 4, 6, 4, 4, 2, 4}),
		buildBoard(5, []int{6, 3, 7, 3, 5, 4, 4, 5, 5, 2, 3, 3, 2, 1, 2, 3, 7, 5, 2, 8, 4, 4, 2, 3, 7}),
		buildBoard(5, []int{1, 4, 1, 4, 6, 2, 2, 4, 2, 4, 5, 5, 2, 4, 6, 8, 2, 2, 7, 1, 5, 5, 6, 1, 1}),
		buildBoard(5, []int{3, 3, 1, 3, 1, 3, 8, 2, 3, 3, 1, 7, 2, 5, 5, 5, 2, 8, 2, 2, 5, 4, 6, 2, 4}),
		buildBoard(5, []int{1, 1, 8, 6, 4, 2, 3, 2, 3, 1, 1, 7, 3, 7, 1, 8, 1, 7, 9, 2, 1, 2, 3, 1, 6}),
		buildBoard(5, []int{4, 4, 4, 3, 3, 2, 4, 2, 1, 3, 4, 2, 2, 6, 1, 5, 9, 1, 1, 3, 5, 1, 2, 6, 2}),
		buildBoard(5, []int{8, 2, 5, 4, 1, 6, 1, 4, 2, 2, 4, 3, 2, 7, 3, 7, 5, 4, 2, 8, 3, 1, 6, 8, 2}),
		buildBoard(5, []int{4, 6, 3, 3, 4, 5, 5, 4, 1, 4, 3, 3, 1, 1, 9, 4, 7, 5, 5, 4, 3, 6, 3, 1, 6}),
		buildBoard(5, []int{9, 8, 2, 4, 6, 1, 1, 9, 9, 1, 6, 1, 1, 2, 5, 8, 2, 4, 6, 1, 1, 6, 2, 1, 4}),
		buildBoard(5, []int{6, 5, 2, 5, 1, 4, 5, 2, 1, 9, 8, 4, 1, 9, 1, 2, 1, 4, 9, 1, 5, 5, 1, 2, 7}),
		buildBoard(5, []int{1, 1, 5, 2, 6, 1, 9, 5, 1, 3, 9, 4, 1, 1, 5, 1, 1, 4, 2, 4, 5, 4, 1, 2, 2}),
		buildBoard(5, []int{1, 9, 8, 2, 2, 2, 4, 2, 2, 8, 1, 5, 2, 2, 2, 6, 2, 5, 3, 6, 4, 4, 2, 4, 2}),
		buildBoard(5, []int{4, 3, 3, 9, 1, 5, 7, 4, 3, 3, 5, 3, 8, 1, 1, 3, 3, 2, 2, 5, 4, 6, 1, 9, 5}),
		buildBoard(5, []int{1, 4, 5, 9, 4, 1, 3, 5, 1, 6, 1, 5, 5, 6, 4, 1, 2, 5, 2, 2, 8, 2, 6, 4, 8}),
		buildBoard(5, []int{2, 9, 3, 2, 6, 8, 1, 2, 3, 4, 1, 1, 1, 7, 1, 3, 1, 7, 3, 1, 7, 9, 2, 8, 8}),
		buildBoard(5, []int{2, 2, 4, 2, 2, 3, 2, 2, 3, 7, 3, 8, 2, 1, 2, 1, 3, 3, 5, 3, 3, 8, 2, 4, 3}),
		buildBoard(5, []int{3, 1, 6, 3, 5, 6, 3, 1, 1, 1, 4, 1, 3, 2, 4, 8, 1, 5, 1, 3, 2, 3, 1, 9, 3}),
		buildBoard(5, []int{3, 2, 5, 7, 3, 1, 8, 5, 7, 3, 3, 4, 4, 2, 6, 3, 8, 2, 9, 2, 4, 4, 2, 1, 2}),
		buildBoard(5, []int{6, 4, 3, 1, 2, 4, 4, 4, 5, 5, 4, 1, 1, 1, 3, 6, 8, 2, 1, 5, 5, 5, 5, 2, 3}),
		buildBoard(5, []int{2, 8, 5, 5, 6, 9, 4, 3, 3, 4, 1, 9, 1, 8, 2, 5, 2, 5, 5, 2, 5, 8, 7, 3, 8}),
		buildBoard(5, []int{4, 2, 4, 1, 3, 1, 2, 5, 1, 5, 1, 6, 2, 7, 2, 3, 3, 3, 1, 1, 3, 1, 8, 2, 9}),
		buildBoard(5, []int{2, 8, 1, 3, 1, 4, 1, 5, 5, 5, 1, 4, 2, 2, 6, 8, 4, 6, 5, 2, 2, 5, 5, 5, 8}),
	}

	g := Game{
		levels: levels,
		level:  0,
		hist:   []Board{levels[0]},
		mouse:  &pkg.CombinedPosControl{},
	}

	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}
