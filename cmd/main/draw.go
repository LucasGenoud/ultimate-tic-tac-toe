package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"GoTicTacToe/lib/engine"
	"GoTicTacToe/lib/graphics"
)

const (
	pieceRadius   = smallCell * 0.28
	pieceWidth    = 6.0
	bigRadius     = miniSize * 0.34
	bigWidth      = 14.0
	cardRadius    = 16.0
	controlRadius = 10.0
)

var newGameButton = rect{x: boardX + boardSize - 132, y: footerY - 24, w: 132, h: 48}

// menuLayout positions the controls of the new game card.
type menuLayout struct {
	card      rect
	opponents [2]rect
	levels    [maxLevel]rect
	start     rect
}

func newMenuLayout() menuLayout {
	card := rect{x: viewWidth/2 - 240, y: boardY + boardSize/2 - 190, w: 480, h: 380}
	inner := card.w - 80
	m := menuLayout{card: card, start: rect{x: card.x + 40, y: card.y + 292, w: inner, h: 52}}
	const gap = 10.0
	for i := range m.opponents {
		w := (inner - gap) / 2
		m.opponents[i] = rect{x: card.x + 40 + float64(i)*(w+gap), y: card.y + 112, w: w, h: 46}
	}
	for i := range m.levels {
		w := (inner - gap*(maxLevel-1)) / maxLevel
		m.levels[i] = rect{x: card.x + 40 + float64(i)*(w+gap), y: card.y + 210, w: w, h: 46}
	}
	return m
}

func resultCard() rect {
	return rect{x: viewWidth/2 - 200, y: boardY + boardSize/2 - 100, w: 400, h: 200}
}

func resultCardButton() rect {
	c := resultCard()
	return rect{x: c.x + 40, y: c.y + c.h - 76, w: c.w - 80, h: 48}
}

func (a *App) Draw(screen *ebiten.Image) {
	screen.Fill(colorBackground)
	c := graphics.NewCanvas(screen, a.viewport)
	a.drawHeader(c)
	a.drawBoard(c)
	a.drawFooter(c)
	for _, p := range a.confetti {
		alpha := clamp01((confettiLifetime - p.age) / 0.6)
		c.FillQuad(p.x, p.y, p.w, p.h, p.angle, graphics.Fade(p.clr, alpha))
	}
	switch a.phase {
	case phaseMenu:
		a.drawMenu(c)
	case phaseOver:
		a.drawResultCard(c)
	}
}

func (a *App) drawHeader(c *graphics.Canvas) {
	c.Text("Ultimate Tic-Tac-Toe", boardX, headerY-16, 30, graphics.Medium, graphics.AlignStart, colorInk)

	status, symbol := a.status()
	x := boardX
	if symbol != engine.Nobody {
		drawSymbol(c, symbol, x+9, headerY+24, 8, 3, 1, playerColor(symbol))
		x += 28
	}
	c.Text(status, x, headerY+24, 18, graphics.Regular, graphics.AlignStart, colorMuted)

	labels := [3]string{engine.O: "Player O", engine.X: "Player X"}
	if a.vsComputer {
		labels[human], labels[computer] = "You", "Computer"
	}
	right := boardX + boardSize
	for _, p := range []engine.Player{engine.X, engine.O} {
		score := fmt.Sprint(a.score[p])
		left := right - math.Max(graphics.TextWidth(labels[p], 14, graphics.Regular), 36+graphics.TextWidth(score, 30, graphics.Medium))
		drawSymbol(c, p, left+11, headerY-14, 10, 4, 1, playerColor(p))
		c.Text(score, right, headerY-16, 30, graphics.Medium, graphics.AlignEnd, colorInk)
		c.Text(labels[p], right, headerY+24, 14, graphics.Regular, graphics.AlignEnd, colorMuted)
		right = left - 36
	}
}

// status describes the state of the game, with the symbol it is about.
func (a *App) status() (string, engine.Player) {
	turn := a.game.Turn()
	switch a.phase {
	case phaseMenu:
		return "Set up a new game", engine.Nobody
	case phaseOver:
		winner := a.game.Winner()
		switch a.outcome() {
		case outcomeDraw:
			return "Draw", engine.Nobody
		case outcomeLoss:
			return "The computer won", winner
		}
		if a.vsComputer {
			return "You won", winner
		}
		return winner.String() + " won", winner
	}
	if a.computerToPlay() {
		return "Computer is thinking" + strings.Repeat(".", int(a.clock*3)%4), turn
	}
	msg := turn.String() + " to play"
	if a.vsComputer {
		msg = "Your turn"
	}
	playable := 0
	for b := range 9 {
		if a.game.BoardPlayable(b) {
			playable++
		}
	}
	if playable > 1 && a.game.LastMove() != engine.NoMove {
		msg += ", any highlighted board"
	}
	return msg, turn
}

func (a *App) drawBoard(c *graphics.Canvas) {
	ended := a.phase == phaseOver
	if ended && a.outcome() == outcomeLoss {
		c.Translate(shakeOffset(a.clock-a.endedAt-outcomeDelay), 0)
		defer c.Translate(0, 0)
	}

	c.FillRect(boardX, boardY+4, boardSize, boardSize, cardRadius+2, colorShadow)
	c.FillRect(boardX, boardY, boardSize, boardSize, cardRadius, colorSurface)
	c.StrokeRect(boardX, boardY, boardSize, boardSize, cardRadius, 1, colorBorder)

	for b := range 9 {
		a.drawMiniBoard(c, b)
	}
	for i := 1; i < 3; i++ {
		at := gridX + float64(i)*bigCell
		c.Line(at, gridY+8, at, gridY+gridSize-8, 5, 1, colorGridMajor)
		c.Line(gridX+8, at-gridX+gridY, gridX+gridSize-8, at-gridX+gridY, 5, 1, colorGridMajor)
	}

	if m := a.hoveredMove(); m != engine.NoMove {
		x, y := cellRect(m).center()
		drawSymbol(c, a.game.Turn(), x, y, pieceRadius, pieceWidth, 1, graphics.Fade(playerColor(a.game.Turn()), 0.22))
	}

	if !ended {
		return
	}
	if a.outcome() != outcomeWin {
		fade := easeOutCubic(progress(a.clock, a.endedAt, outcomeDelay, fadeDuration))
		c.FillRect(gridX, gridY, gridSize, gridSize, 0, graphics.Fade(colorSurface, 0.6*fade))
	}
	if line, ok := a.game.WinningLine(); ok {
		x0, y0 := bigCellCenter(line[0])
		x1, y1 := bigCellCenter(line[2])
		x0, y0, x1, y1 = extend(x0, y0, x1, y1, bigCell*0.38)
		t := easeInOutCubic(progress(a.clock, a.endedAt, endLineDelay, endLineDuration))
		c.Line(x0, y0, x1, y1, 24, t, colorSurface)
		c.Line(x0, y0, x1, y1, 14, t, playerColor(a.game.Winner()))
	}
}

func (a *App) drawMiniBoard(c *graphics.Canvas, board int) {
	r := miniBoardRect(board)
	if h := a.highlight[board]; h > 0 {
		c.FillRect(r.x-8, r.y-8, r.w+16, r.h+16, controlRadius, graphics.Fade(playerColor(a.game.Turn()), 0.09*h))
	}
	if last := a.game.LastMove(); last != engine.NoMove && last.Board() == board && a.phase == phasePlaying {
		cell := cellRect(last)
		c.FillRect(cell.x+5, cell.y+5, cell.w-10, cell.h-10, 8, graphics.Fade(colorInk, 0.05))
	}
	for i := 1; i < 3; i++ {
		at := float64(i) * smallCell
		c.Line(r.x+at, r.y+4, r.x+at, r.y+r.h-4, 2, 1, colorGridMinor)
		c.Line(r.x+4, r.y+at, r.x+r.w-4, r.y+at, 2, 1, colorGridMinor)
	}

	winner := a.game.BoardWinner(board)
	symbolStart := boardDelay + strikeDuration
	pieceAlpha := 1.0
	switch {
	case winner != engine.Nobody:
		pieceAlpha = 1 - 0.85*easeOutCubic(progress(a.clock, a.decidedAt[board], symbolStart, bigSymbolDuration))
	case a.game.BoardDone(board):
		pieceAlpha = 1 - 0.6*easeOutCubic(progress(a.clock, a.decidedAt[board], boardDelay, fadeDuration))
	}

	for cell := range 9 {
		m := engine.MoveAt(board, cell)
		if p := a.game.Cell(m); p != engine.Nobody {
			x, y := cellRect(m).center()
			t := easeOutCubic(progress(a.clock, a.placedAt[m], 0, pieceDuration))
			drawSymbol(c, p, x, y, pieceRadius, pieceWidth, t, graphics.Fade(playerColor(p), pieceAlpha))
		}
	}

	if line, ok := a.game.BoardLine(board); ok {
		x0, y0 := cellRect(engine.MoveAt(board, line[0])).center()
		x1, y1 := cellRect(engine.MoveAt(board, line[2])).center()
		x0, y0, x1, y1 = extend(x0, y0, x1, y1, smallCell*0.32)
		t := easeInOutCubic(progress(a.clock, a.decidedAt[board], boardDelay, strikeDuration))
		c.Line(x0, y0, x1, y1, 5, t, graphics.Fade(playerColor(winner), pieceAlpha))

		cx, cy := r.center()
		t = easeOutCubic(progress(a.clock, a.decidedAt[board], symbolStart, bigSymbolDuration))
		drawSymbol(c, winner, cx, cy, bigRadius, bigWidth, t, playerColor(winner))
	}
}

func (a *App) drawFooter(c *graphics.Canvas) {
	line1, line2 := "Two players on one device", "Hold R to reset the score"
	if a.vsComputer {
		line1 = fmt.Sprintf("Computer level %d, %d s per move", a.level, a.level)
		if s := a.lastSearch; s.Workers > 0 {
			line2 = fmt.Sprintf("Last move: %s simulations on %d threads, %.0f%% win estimate",
				thousands(s.Simulations), s.Workers, s.WinRate*100)
		}
	}
	if held := ebiten.IsKeyPressed(ebiten.KeyR); held {
		line2 = "Keep holding R to reset the score"
	}
	c.Text(line1, boardX, footerY-11, 16, graphics.Regular, graphics.AlignStart, colorInk)
	c.Text(line2, boardX, footerY+13, 14, graphics.Regular, graphics.AlignStart, colorMuted)

	if a.phase != phaseMenu {
		a.drawButton(c, newGameButton, "New game", false, true)
	}
}

func (a *App) drawMenu(c *graphics.Canvas) {
	c.FillRect(boardX, boardY, boardSize, boardSize, cardRadius, graphics.Fade(colorBackground, 0.8))
	m := newMenuLayout()
	drawCard(c, m.card, 1)
	x := m.card.x + 40
	c.Text("New game", x, m.card.y+52, 26, graphics.Medium, graphics.AlignStart, colorInk)

	c.Text("Opponent", x, m.opponents[0].y-20, 15, graphics.Regular, graphics.AlignStart, colorMuted)
	a.drawButton(c, m.opponents[0], "Computer", a.vsComputer, true)
	a.drawButton(c, m.opponents[1], "Friend", !a.vsComputer, true)

	c.Text("Computer level (seconds per move)", x, m.levels[0].y-20, 15, graphics.Regular, graphics.AlignStart, colorMuted)
	for i, r := range m.levels {
		a.drawButton(c, r, fmt.Sprint(i+1), a.vsComputer && a.level == i+1, a.vsComputer)
	}

	a.drawButton(c, m.start, "Start", true, true)
	c.Text("Space to start, A to switch opponent, 1 to 5 to set the level",
		viewWidth/2, m.card.y+m.card.h+30, 14, graphics.Regular, graphics.AlignCenter, colorMuted)
}

func (a *App) drawResultCard(c *graphics.Canvas) {
	t := easeOutCubic(progress(a.clock, a.endedAt, cardDelay, cardDuration))
	if t == 0 {
		return
	}
	r := resultCard()
	r.y += (1 - t) * 24
	drawCard(c, r, t)

	title := map[outcome]string{outcomeWin: "You won", outcomeLoss: "You lost", outcomeDraw: "Draw"}[a.outcome()]
	if !a.vsComputer && a.outcome() == outcomeWin {
		title = a.game.Winner().String() + " won"
	}
	c.Text(title, viewWidth/2, r.y+50, 30, graphics.Medium, graphics.AlignCenter, graphics.Fade(colorInk, t))
	score := fmt.Sprintf("O %d  ·  X %d", a.score[engine.O], a.score[engine.X])
	c.Text(score, viewWidth/2, r.y+88, 16, graphics.Regular, graphics.AlignCenter, graphics.Fade(colorMuted, t))

	button := resultCardButton()
	button.y = r.y + r.h - 76
	if t == 1 {
		a.drawButton(c, button, "Play again", true, true)
	}
}

func drawCard(c *graphics.Canvas, r rect, alpha float64) {
	c.FillRect(r.x, r.y+6, r.w, r.h, cardRadius+2, graphics.Fade(colorShadow, alpha*1.5))
	c.FillRect(r.x, r.y, r.w, r.h, cardRadius, graphics.Fade(colorSurface, alpha))
	c.StrokeRect(r.x, r.y, r.w, r.h, cardRadius, 1, graphics.Fade(colorBorder, alpha))
}

// drawButton draws a filled button when selected, an outlined one otherwise.
func (a *App) drawButton(c *graphics.Canvas, r rect, label string, selected, enabled bool) {
	hovered := enabled && r.contains(a.pointerX, a.pointerY)
	fill, border, ink := colorSurface, colorBorder, colorInk
	switch {
	case !enabled:
		ink = graphics.Fade(colorMuted, 0.5)
	case selected:
		fill, border, ink = colorInk, colorInk, colorSurface
		if hovered {
			fill = graphics.Mix(colorInk, colorSurface, 0.15)
		}
	case hovered:
		fill, border = colorBackground, colorGridMinor
	}
	c.FillRect(r.x, r.y, r.w, r.h, controlRadius, fill)
	c.StrokeRect(r.x, r.y, r.w, r.h, controlRadius, 1, border)
	x, y := r.center()
	c.Text(label, x, y, 16, graphics.Medium, graphics.AlignCenter, ink)
}

func drawSymbol(c *graphics.Canvas, p engine.Player, x, y, radius, width, progress float64, clr color.Color) {
	if p == engine.O {
		c.Nought(x, y, radius, width, progress, clr)
	} else {
		c.Cross(x, y, radius*1.1, width, progress, clr)
	}
}

// extend lengthens the segment by margin at both ends.
func extend(x0, y0, x1, y1, margin float64) (float64, float64, float64, float64) {
	l := math.Hypot(x1-x0, y1-y0)
	dx, dy := (x1-x0)/l*margin, (y1-y0)/l*margin
	return x0 - dx, y0 - dy, x1 + dx, y1 + dy
}

// thousands formats n with comma separators.
func thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
