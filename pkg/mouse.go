package pkg

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Pos struct {
	X, Y float32
}

func (p Pos) String() string {
	return fmt.Sprintf("(%d, %d)", int(p.X), int(p.Y))
}

type PosControl interface {
	IsDragging() (bool, Pos, Pos)
	JustFinishedDragging() (bool, Pos, Pos)
	Update() error
}

type MouseControl struct {
	sp                   Pos
	dragging             bool
	justFinishedDragging bool
}

func (m *MouseControl) cursorPos() Pos {
	x, y := ebiten.CursorPosition()
	return Pos{float32(x), float32(y)}
}

func (m *MouseControl) Update() error {
	m.justFinishedDragging = false

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		m.dragging, m.justFinishedDragging = false, true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		m.dragging, m.sp = true, m.cursorPos()
	}
	return nil
}

func (m *MouseControl) IsDragging() (bool, Pos, Pos) {
	return m.dragging, m.sp, m.cursorPos()
}

func (m *MouseControl) JustFinishedDragging() (bool, Pos, Pos) {
	return m.justFinishedDragging, m.sp, m.cursorPos()
}

type TouchControl struct {
	drugging             bool
	id                   ebiten.TouchID
	sp, fp               Pos
	justFinishedDragging bool
}

func (t *TouchControl) IsDragging() (bool, Pos, Pos) {
	return t.drugging, t.sp, t.fp
}

func (t *TouchControl) JustFinishedDragging() (bool, Pos, Pos) {
	return t.justFinishedDragging, t.sp, t.fp
}

func (t *TouchControl) touchPos() Pos {
	x, y := ebiten.TouchPosition(t.id)
	return Pos{float32(x), float32(y)}
}

func (t *TouchControl) Update() error {
	t.justFinishedDragging = false
	if !t.drugging {
		touches := inpututil.AppendJustPressedTouchIDs(nil)
		if len(touches) == 0 {
			return nil
		}
		t.drugging, t.id, t.sp, t.fp = true, touches[0], t.touchPos(), t.touchPos()
	}

	if inpututil.IsTouchJustReleased(t.id) {
		t.drugging, t.justFinishedDragging = false, true
	} else {
		t.fp = t.touchPos()
	}
	return nil
}

type CombinedPosControl struct {
	m MouseControl
	t TouchControl
}

func (c *CombinedPosControl) IsDragging() (bool, Pos, Pos) {
	if ok, s, f := c.m.IsDragging(); ok {
		return true, s, f
	}
	return c.t.IsDragging()
}

func (c *CombinedPosControl) JustFinishedDragging() (bool, Pos, Pos) {
	if ok, s, f := c.m.JustFinishedDragging(); ok {
		return true, s, f
	}
	return c.t.JustFinishedDragging()
}

func (c *CombinedPosControl) Update() error {
	c.m.Update()
	c.t.Update()
	return nil
}
