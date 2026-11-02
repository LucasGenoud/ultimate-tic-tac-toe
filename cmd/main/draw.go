package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"GoTicTacToe/lib/ai"
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

func resultCard() rect {
	return rect{x: viewWidth/2 - 200, y: boardY + boardSize/2 - 100, w: 400, h: 200}
}

func resultCardButton() rect {
	c := resultCard()
	return rect{x: c.x + 40, y: c.y + c.h - 76, w: c.w - 80, h: 48}
}

func (a *App) Draw(screen *ebiten.Image) {
	p := a.pal()
	screen.Fill(p.background)
	c := graphics.NewCanvas(screen, a.viewport)
	a.drawHeader(c)
	a.drawBoard(c)
	a.drawStats(c)
	a.drawFooter(c)
	for _, piece := range a.confetti {
		alpha := clamp01((confettiLifetime - piece.age) / 0.6)
		c.FillQuad(piece.x, piece.y, piece.w, piece.h, piece.angle, graphics.Fade(piece.clr, alpha))
	}
	switch a.phase {
	case phaseMenu:
		a.drawMenu(c)
	case phaseOver:
		a.drawResultCard(c)
	}
	if a.showRules {
		a.drawRules(c)
	}
}

func (a *App) drawHeader(c *graphics.Canvas) {
	p := a.pal()
	c.Text("Ultimate Tic-Tac-Toe", boardX, titleY, 30, graphics.Medium, graphics.AlignStart, p.ink)

	status, symbol := a.status()
	x := boardX
	if symbol != engine.Nobody {
		a.drawSymbol(c, symbol, x+9, statusY, 8, 3, 1, p.player(symbol))
		x += 28
	}
	c.Text(status, x, statusY, 18, graphics.Regular, graphics.AlignStart, p.muted)

	labels := [3]string{engine.O: "Player O", engine.X: "Player X"}
	if a.settings.VsComputer {
		labels[human], labels[computer] = "You", "Computer"
	}
	right := boardX + boardSize
	for _, pl := range []engine.Player{engine.X, engine.O} {
		score := fmt.Sprint(a.settings.Score[pl])
		left := right - math.Max(graphics.TextWidth(labels[pl], 14, graphics.Regular), 36+graphics.TextWidth(score, 30, graphics.Medium))
		a.drawSymbol(c, pl, left+11, titleY, 10, 4, 1, p.player(pl))
		c.Text(score, right, titleY, 30, graphics.Medium, graphics.AlignEnd, p.ink)
		c.Text(labels[pl], right, statusY, 14, graphics.Regular, graphics.AlignEnd, p.muted)
		right = left - 36
	}
}

// status describes the state of the game, with the symbol it is about.
func (a *App) status() (string, engine.Player) {
	turn := a.game.Turn()
	dots := strings.Repeat(".", int(a.clock*3)%4)
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyR):
		return "Keep holding R to reset the score", engine.Nobody
	case a.phase == phaseMenu:
		return "Set up a new game", engine.Nobody
	case a.phase == phaseOver:
		winner := a.game.Winner()
		switch a.outcome() {
		case outcomeDraw:
			return "Draw", engine.Nobody
		case outcomeLoss:
			return "The computer won", winner
		}
		if a.settings.VsComputer {
			return "You won", winner
		}
		return winner.String() + " won", winner
	case a.computerToPlay():
		return "Computer is thinking" + dots, turn
	case a.request != nil:
		return "Looking for a good move" + dots, turn
	case a.hint != engine.NoMove:
		return "Hint: the outlined cell, " + strings.ToLower(winChance(a.hintResult)) + " win chance", turn
	}
	msg := turn.String() + " to play"
	if a.settings.VsComputer {
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
	p := a.pal()
	ended := a.phase == phaseOver
	if ended && a.outcome() == outcomeLoss {
		c.Translate(shakeOffset(a.clock-a.endedAt-outcomeDelay), 0)
		defer c.Translate(0, 0)
	}

	c.FillRect(boardX, boardY+4, boardSize, boardSize, cardRadius+2, p.shadow)
	c.FillRect(boardX, boardY, boardSize, boardSize, cardRadius, p.surface)
	c.StrokeRect(boardX, boardY, boardSize, boardSize, cardRadius, 1, p.border)

	for b := range 9 {
		a.drawMiniBoard(c, b)
	}
	for i := 1; i < 3; i++ {
		at := float64(i) * bigCell
		c.Line(gridX+at, gridY+8, gridX+at, gridY+gridSize-8, 5, 1, p.gridMajor)
		c.Line(gridX+8, gridY+at, gridX+gridSize-8, gridY+at, 5, 1, p.gridMajor)
	}

	turn := a.game.Turn()
	if a.hint != engine.NoMove {
		r := cellRect(a.hint)
		pulse := 0.55 + 0.35*math.Sin(a.clock*5)
		c.StrokeRect(r.x+4, r.y+4, r.w-8, r.h-8, 10, 3, graphics.Fade(p.player(turn), pulse))
		x, y := r.center()
		a.drawSymbol(c, turn, x, y, pieceRadius, pieceWidth, 1, graphics.Fade(p.player(turn), 0.3))
	}
	if m := a.hoveredMove(); m != engine.NoMove {
		x, y := cellRect(m).center()
		a.drawSymbol(c, turn, x, y, pieceRadius, pieceWidth, 1, graphics.Fade(p.player(turn), 0.22))
	}

	if !ended {
		return
	}
	if a.outcome() != outcomeWin {
		fade := easeOutCubic(progress(a.clock, a.endedAt, outcomeDelay, fadeDuration))
		c.FillRect(gridX, gridY, gridSize, gridSize, 0, graphics.Fade(p.surface, 0.6*fade))
	}
	if line, ok := a.game.WinningLine(); ok {
		x0, y0 := bigCellCenter(line[0])
		x1, y1 := bigCellCenter(line[2])
		x0, y0, x1, y1 = extend(x0, y0, x1, y1, bigCell*0.38)
		t := easeInOutCubic(progress(a.clock, a.endedAt, endLineDelay, endLineDuration))
		c.Line(x0, y0, x1, y1, 24, t, p.surface)
		c.Line(x0, y0, x1, y1, 14, t, p.player(a.game.Winner()))
	}
}

func (a *App) drawMiniBoard(c *graphics.Canvas, board int) {
	p := a.pal()
	r := miniBoardRect(board)
	if h := a.highlight[board]; h > 0 {
		c.FillRect(r.x-8, r.y-8, r.w+16, r.h+16, controlRadius, graphics.Fade(p.player(a.game.Turn()), 0.1*h))
	}
	if last := a.game.LastMove(); last != engine.NoMove && last.Board() == board && a.phase == phasePlaying {
		cell := cellRect(last)
		c.FillRect(cell.x+5, cell.y+5, cell.w-10, cell.h-10, 8, graphics.Fade(p.ink, 0.06))
	}
	for i := 1; i < 3; i++ {
		at := float64(i) * smallCell
		c.Line(r.x+at, r.y+4, r.x+at, r.y+r.h-4, 2, 1, p.gridMinor)
		c.Line(r.x+4, r.y+at, r.x+r.w-4, r.y+at, 2, 1, p.gridMinor)
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
		if pl := a.game.Cell(m); pl != engine.Nobody {
			x, y := cellRect(m).center()
			t := easeOutCubic(progress(a.clock, a.placedAt[m], 0, pieceDuration))
			a.drawSymbol(c, pl, x, y, pieceRadius, pieceWidth, t, graphics.Fade(p.player(pl), pieceAlpha))
		}
	}

	if line, ok := a.game.BoardLine(board); ok {
		x0, y0 := cellRect(engine.MoveAt(board, line[0])).center()
		x1, y1 := cellRect(engine.MoveAt(board, line[2])).center()
		x0, y0, x1, y1 = extend(x0, y0, x1, y1, smallCell*0.32)
		t := easeInOutCubic(progress(a.clock, a.decidedAt[board], boardDelay, strikeDuration))
		c.Line(x0, y0, x1, y1, 5, t, graphics.Fade(p.player(winner), pieceAlpha))

		cx, cy := r.center()
		t = easeOutCubic(progress(a.clock, a.decidedAt[board], symbolStart, bigSymbolDuration))
		a.drawSymbol(c, winner, cx, cy, bigRadius, bigWidth, t, p.player(winner))
	}
}

// drawStats shows what the computer found: live while it thinks, then for its last move.
func (a *App) drawStats(c *graphics.Canvas) {
	p := a.pal()
	stat := func(x float64, label, value string) {
		c.Text(label, x, statsLabelY, 13, graphics.Regular, graphics.AlignStart, p.muted)
		c.Text(value, x, statsValueY, 20, graphics.Medium, graphics.AlignStart, p.ink)
	}
	if !a.settings.VsComputer {
		stat(boardX, "Mode", "Two players on one device")
		return
	}

	res := a.lastSearch
	thinking := a.request != nil && a.request.kind == requestMove
	if thinking {
		res = a.live
	}
	simulations := "-"
	if res.Simulations > 0 {
		simulations = thousands(res.Simulations)
	}
	threads := fmt.Sprintf("%d of %d", a.threads(), a.searcher.MaxThreads())
	if a.searcher.Fallback() {
		threads = "1 (no workers)"
	}
	stat(boardX, "Level", a.level().name)
	stat(boardX+170, "Simulations", simulations)
	stat(boardX+360, "Computer win chance", winChance(res))
	stat(boardX+560, "Threads", threads)

	if thinking {
		limits := a.request.limits
		t := (a.clock - a.request.started) / limits.Duration.Seconds()
		if limits.Simulations > 0 {
			t = float64(a.live.Simulations) / float64(limits.Simulations)
		}
		c.Line(boardX, barY, boardX+boardSize, barY, 3, 1, p.border)
		c.Line(boardX, barY, boardX+boardSize, barY, 3, clamp01(t), p.ink)
	}
}

// winChance formats the estimated chance of winning of a search result.
func winChance(r ai.Result) string {
	switch {
	case r.Threads == 0:
		return "-"
	case r.Proof == ai.Win:
		return "Forced win"
	case r.Proof == ai.Loss:
		return "Forced loss"
	}
	return fmt.Sprintf("%.0f%%", r.WinRate*100)
}

func (a *App) drawFooter(c *graphics.Canvas) {
	if a.phase == phaseMenu {
		return
	}
	a.drawButton(c, hintButton, "Hint", false, a.canHint())
	a.drawButton(c, undoButton, "Undo", false, a.canUndo())
	a.drawButton(c, newGameButton, "New game", false, true)
}

func (a *App) drawResultCard(c *graphics.Canvas) {
	p := a.pal()
	t := easeOutCubic(progress(a.clock, a.endedAt, cardDelay, cardDuration))
	if t == 0 {
		return
	}
	r := resultCard()
	r.y += (1 - t) * 24
	a.drawCard(c, r, t)

	title := map[outcome]string{outcomeWin: "You won", outcomeLoss: "You lost", outcomeDraw: "Draw"}[a.outcome()]
	if !a.settings.VsComputer && a.outcome() == outcomeWin {
		title = a.game.Winner().String() + " won"
	}
	c.Text(title, viewWidth/2, r.y+50, 30, graphics.Medium, graphics.AlignCenter, graphics.Fade(p.ink, t))
	score := fmt.Sprintf("O %d  ·  X %d", a.settings.Score[engine.O], a.settings.Score[engine.X])
	c.Text(score, viewWidth/2, r.y+88, 16, graphics.Regular, graphics.AlignCenter, graphics.Fade(p.muted, t))
	if t == 1 {
		a.drawButton(c, resultCardButton(), "Play again", true, true)
	}
}

func (a *App) drawCard(c *graphics.Canvas, r rect, alpha float64) {
	p := a.pal()
	c.FillRect(r.x, r.y+6, r.w, r.h, cardRadius+2, graphics.Fade(p.shadow, alpha*1.5))
	c.FillRect(r.x, r.y, r.w, r.h, cardRadius, graphics.Fade(p.surface, alpha))
	c.StrokeRect(r.x, r.y, r.w, r.h, cardRadius, 1, graphics.Fade(p.border, alpha))
}

// drawButton draws a filled button when selected, an outlined one otherwise.
func (a *App) drawButton(c *graphics.Canvas, r rect, label string, selected, enabled bool) {
	p := a.pal()
	hovered := enabled && r.contains(a.pointerX, a.pointerY)
	fill, border, ink := p.surface, p.border, p.ink
	switch {
	case !enabled:
		ink = graphics.Fade(p.muted, 0.5)
	case selected:
		fill, border, ink = p.ink, p.ink, p.surface
		if hovered {
			fill = graphics.Mix(p.ink, p.surface, 0.15)
		}
	case hovered:
		fill, border = p.background, p.gridMinor
	}
	c.FillRect(r.x, r.y, r.w, r.h, controlRadius, fill)
	c.StrokeRect(r.x, r.y, r.w, r.h, controlRadius, 1, border)
	x, y := r.center()
	c.Text(label, x, y, 15, graphics.Medium, graphics.AlignCenter, ink)
}

func (a *App) drawSymbol(c *graphics.Canvas, pl engine.Player, x, y, radius, width, progress float64, clr color.Color) {
	if pl == engine.O {
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
