package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"GoTicTacToe/lib/graphics"
)

// menuLayout positions the controls of the new game card.
type menuLayout struct {
	card      rect
	rules     rect
	opponents []rect
	levels    []rect
	first     []rect // you or O, the computer or X, alternate
	slider    rect
	sound     []rect // on, off
	theme     []rect // light, dark
	start     rect
}

// firstMoveOrder maps the first move buttons to their setting.
var firstMoveOrder = [3]firstMove{firstO, firstX, firstAlternate}

func newMenuLayout() menuLayout {
	card := rect{x: viewWidth/2 - 260, y: boardY + 68, w: 520, h: 584}
	row := func(y, h float64, n int) []rect {
		return rect{x: card.x + 36, y: card.y + y, w: card.w - 72, h: h}.row(n, 10)
	}
	halves := row(438, 44, 2)
	return menuLayout{
		card:      card,
		rules:     rect{x: card.x + card.w - 36 - 120, y: card.y + 28, w: 120, h: 36},
		opponents: row(102, 44, 2),
		levels:    row(184, 44, len(levels)),
		first:     row(292, 44, 3),
		slider:    row(370, 28, 1)[0],
		sound:     halves[0].row(2, 8),
		theme:     halves[1].row(2, 8),
		start:     row(506, 52, 1)[0],
	}
}

func hit(rects []rect, x, y float64) int {
	for i, r := range rects {
		if r.contains(x, y) {
			return i
		}
	}
	return -1
}

func (a *App) updateMenu(clicked bool) {
	m := newMenuLayout()
	x, y := a.pointerX, a.pointerY

	if a.draggingSlider {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || len(ebiten.AppendTouchIDs(nil)) > 0 {
			a.setThreadsAt(m.slider, x)
		} else {
			a.draggingSlider = false
			a.save()
		}
		return
	}

	for i, key := 0, ebiten.Key1; i < len(levels); i, key = i+1, key+1 {
		if inpututil.IsKeyJustPressed(key) {
			a.settings.Level = i
			a.save()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		a.setOpponent(!a.settings.VsComputer)
		a.save()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		a.startGame()
		return
	}
	if !clicked {
		return
	}

	switch {
	case m.rules.contains(x, y):
		a.showRules = true
	case hit(m.opponents, x, y) >= 0:
		a.setOpponent(hit(m.opponents, x, y) == 0)
	case hit(m.levels, x, y) >= 0 && a.settings.VsComputer:
		a.settings.Level = hit(m.levels, x, y)
	case hit(m.first, x, y) >= 0:
		a.settings.FirstMove = firstMoveOrder[hit(m.first, x, y)]
	case a.sliderHitArea(m.slider).contains(x, y) && a.searcher.MaxThreads() > 1:
		a.draggingSlider = true
		a.setThreadsAt(m.slider, x)
	case hit(m.sound, x, y) >= 0:
		a.setSound(hit(m.sound, x, y) == 0)
	case hit(m.theme, x, y) >= 0:
		a.settings.Dark = hit(m.theme, x, y) == 1
	case m.start.contains(x, y):
		a.startGame()
	}
	a.save()
}

func (a *App) sliderHitArea(track rect) rect {
	return rect{x: track.x - 12, y: track.y - 10, w: track.w + 24, h: track.h + 20}
}

func (a *App) setThreadsAt(track rect, x float64) {
	n := a.searcher.MaxThreads()
	t := clamp01((x - track.x) / track.w)
	a.settings.Threads = 1 + int(math.Round(t*float64(n-1)))
}

func (a *App) menuControlAt(x, y float64) bool {
	m := newMenuLayout()
	return m.rules.contains(x, y) || hit(m.opponents, x, y) >= 0 || hit(m.first, x, y) >= 0 ||
		(a.settings.VsComputer && hit(m.levels, x, y) >= 0) ||
		(a.searcher.MaxThreads() > 1 && a.sliderHitArea(m.slider).contains(x, y)) ||
		hit(m.sound, x, y) >= 0 || hit(m.theme, x, y) >= 0 || m.start.contains(x, y)
}

func (a *App) drawMenu(c *graphics.Canvas) {
	p := a.pal()
	c.FillRect(boardX, boardY, boardSize, boardSize, cardRadius, graphics.Fade(p.background, 0.8))
	m := newMenuLayout()
	a.drawCard(c, m.card, 1)
	left := m.card.x + 36
	label := func(s string, x, y float64) {
		c.Text(s, x, y, 15, graphics.Regular, graphics.AlignStart, p.muted)
	}

	c.Text("New game", left, m.card.y+46, 26, graphics.Medium, graphics.AlignStart, p.ink)
	a.drawButton(c, m.rules, "How to play", false, true)

	label("Opponent", left, m.opponents[0].y-16)
	vs := a.settings.VsComputer
	a.drawButton(c, m.opponents[0], "Computer", vs, true)
	a.drawButton(c, m.opponents[1], "Friend", !vs, true)

	label("Computer level", left, m.levels[0].y-16)
	for i, r := range m.levels {
		a.drawButton(c, r, levels[i].name, vs && a.settings.Level == i, vs)
	}
	about := "The computer is off when playing with a friend"
	if vs {
		about = a.level().about
	}
	c.Text(about, left, m.levels[0].y+m.levels[0].h+18, 14, graphics.Regular, graphics.AlignStart, p.muted)

	label("First move", left, m.first[0].y-16)
	names := [3]string{"O", "X", "Alternate"}
	if vs {
		names[0], names[1] = "You", "Computer"
	}
	for i, r := range m.first {
		a.drawButton(c, r, names[i], a.settings.FirstMove == firstMoveOrder[i], true)
	}

	a.drawSlider(c, m.slider)

	label("Sound", left, m.sound[0].y-16)
	a.drawButton(c, m.sound[0], "On", a.settings.Sound, true)
	a.drawButton(c, m.sound[1], "Off", !a.settings.Sound, true)
	label("Theme", m.theme[0].x, m.theme[0].y-16)
	a.drawButton(c, m.theme[0], "Light", !a.settings.Dark, true)
	a.drawButton(c, m.theme[1], "Dark", a.settings.Dark, true)

	a.drawButton(c, m.start, "Start", true, true)
	c.Text("Space starts, A switches opponent, 1 to 5 set the level, M sound, D theme",
		viewWidth/2, m.card.y+m.card.h+26, 14, graphics.Regular, graphics.AlignCenter, p.muted)
}

func (a *App) drawSlider(c *graphics.Canvas, track rect) {
	p := a.pal()
	n := a.searcher.MaxThreads()
	value := fmt.Sprintf("%d of %d", a.threads(), n)
	if a.searcher.Fallback() {
		value = "1, Web Workers unavailable"
	}
	c.Text("Threads used by the computer", track.x, track.y-16, 15, graphics.Regular, graphics.AlignStart, p.muted)
	c.Text(value, track.x+track.w, track.y-16, 15, graphics.Medium, graphics.AlignEnd, p.ink)

	_, cy := track.center()
	t := 1.0
	if n > 1 {
		t = float64(a.threads()-1) / float64(n-1)
	}
	knob := track.x + t*track.w
	c.Line(track.x, cy, track.x+track.w, cy, 4, 1, p.border)
	c.Line(track.x, cy, knob, cy, 4, 1, p.ink)
	c.FillCircle(knob, cy, 11, p.ink)
	c.FillCircle(knob, cy, 4, p.surface)
}

// rulesCard is the "How to play" card.
func rulesCard() rect {
	return rect{x: viewWidth/2 - 300, y: boardY + 110, w: 600, h: 470}
}

func rulesButton() rect {
	r := rulesCard()
	return rect{x: r.x + 36, y: r.y + r.h - 84, w: r.w - 72, h: 52}
}

var rules = []string{
	"The board is made of nine small tic-tac-toe boards.",
	"Win a small board with three in a row inside it,",
	"and the game with three small boards in a row.",
	"",
	"The cell you play sends your opponent to the matching board:",
	"a top right cell sends them to the top right board.",
	"If that board is already decided, they may play in any open board.",
	"Boards where the next move can be played are highlighted.",
}

func (a *App) updateRules(clicked bool) {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEscape) || (clicked && rulesButton().contains(a.pointerX, a.pointerY)) {
		a.showRules = false
		a.settings.SeenRules = true
		a.save()
	}
}

func (a *App) drawRules(c *graphics.Canvas) {
	p := a.pal()
	c.FillRect(0, 0, viewWidth, viewHeight, 0, graphics.Fade(p.background, 0.7))
	r := rulesCard()
	a.drawCard(c, r, 1)
	c.Text("How to play", r.x+36, r.y+50, 26, graphics.Medium, graphics.AlignStart, p.ink)
	for i, line := range rules {
		c.Text(line, r.x+36, r.y+104+float64(i)*32, 17, graphics.Regular, graphics.AlignStart, p.ink)
	}
	a.drawButton(c, rulesButton(), "Got it", true, true)
}
