package main

import (
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"GoTicTacToe/lib/ai"
	"GoTicTacToe/lib/engine"
	"GoTicTacToe/lib/graphics"
)

type phase int

const (
	phaseMenu phase = iota
	phasePlaying
	phaseOver
)

type outcome int

const (
	outcomeWin outcome = iota
	outcomeLoss
	outcomeDraw
)

const (
	human        = engine.O
	computer     = engine.X
	defaultLevel = 2
	maxLevel     = 5
)

type aiReply struct {
	match  int
	result ai.Result
}

// App is the ebiten game: it owns the match, the menus and the animations.
type App struct {
	game       engine.Game
	phase      phase
	vsComputer bool
	level      int // seconds the computer thinks per move
	nextFirst  engine.Player
	score      [3]int // wins indexed by player, draws under engine.Nobody
	match      int    // incremented for every game so stale computer replies are ignored

	thinking   bool
	aiReplies  chan aiReply
	lastSearch ai.Result

	clock      float64 // seconds since launch, drives every animation
	placedAt   [81]float64
	decidedAt  [9]float64
	endedAt    float64
	highlight  [9]float64 // eased highlight of the playable boards
	confetti   confetti
	celebrated bool

	viewport           graphics.Viewport
	pointerX, pointerY float64
}

func NewApp() *App {
	first := engine.O
	if rand.IntN(2) == 0 {
		first = engine.X
	}
	return &App{
		game:       engine.New(first),
		vsComputer: true,
		level:      defaultLevel,
		nextFirst:  first,
		aiReplies:  make(chan aiReply, 1),
	}
}

// LayoutF renders at the device resolution so text and shapes stay sharp on high density screens.
func (a *App) LayoutF(width, height float64) (float64, float64) {
	scale := ebiten.Monitor().DeviceScaleFactor()
	a.viewport = graphics.Fit(width*scale, height*scale, viewWidth, viewHeight)
	return width * scale, height * scale
}

// Layout is required by ebiten.Game but never called since App implements LayoutF.
func (a *App) Layout(width, height int) (int, int) {
	w, h := a.LayoutF(float64(width), float64(height))
	return int(w), int(h)
}

func (a *App) Update() error {
	dt := 1 / float64(ebiten.TPS())
	a.clock += dt
	a.confetti.update(dt)
	a.pointerX, a.pointerY = a.viewport.ToVirtual(ebiten.CursorPositionF())
	clicked := a.pointerPressed()

	if err := a.handleGlobalKeys(); err != nil {
		return err
	}
	if clicked && a.phase != phaseMenu && newGameButton.contains(a.pointerX, a.pointerY) {
		a.openMenu()
		return nil
	}

	switch a.phase {
	case phaseMenu:
		a.updateMenu(clicked)
	case phasePlaying:
		a.updatePlaying(clicked)
	case phaseOver:
		a.updateOver(clicked)
	}
	for b := range a.highlight {
		target := 0.0
		if a.phase == phasePlaying && a.game.BoardPlayable(b) {
			target = 1
		}
		a.highlight[b] = approach(a.highlight[b], target, dt, highlightDuration)
	}
	a.updateCursor()
	return nil
}

// pointerPressed reports a mouse click or a touch, moving the pointer to the touch position.
func (a *App) pointerPressed() bool {
	if touches := inpututil.AppendJustPressedTouchIDs(nil); len(touches) > 0 {
		a.pointerX, a.pointerY = a.viewport.ToVirtual(ebiten.TouchPositionF(touches[0]))
		return true
	}
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

func (a *App) handleGlobalKeys() error {
	hold := ebiten.TPS() // one second
	if inpututil.KeyPressDuration(ebiten.KeyR) == hold {
		a.score = [3]int{}
		a.openMenu()
	}
	if runtime.GOOS != "js" && inpututil.KeyPressDuration(ebiten.KeyEscape) == hold {
		return ebiten.Termination
	}
	return nil
}

func (a *App) updateMenu(clicked bool) {
	menu := newMenuLayout()
	for i, key := 0, ebiten.Key1; i < maxLevel; i, key = i+1, key+1 {
		if inpututil.IsKeyJustPressed(key) || (clicked && a.vsComputer && menu.levels[i].contains(a.pointerX, a.pointerY)) {
			a.level = i + 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		a.setOpponent(!a.vsComputer)
	}
	if clicked && menu.opponents[0].contains(a.pointerX, a.pointerY) {
		a.setOpponent(true)
	}
	if clicked && menu.opponents[1].contains(a.pointerX, a.pointerY) {
		a.setOpponent(false)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || (clicked && menu.start.contains(a.pointerX, a.pointerY)) {
		a.startGame()
	}
}

// setOpponent switches between playing the computer and a friend, starting a new score.
func (a *App) setOpponent(vsComputer bool) {
	if a.vsComputer != vsComputer {
		a.vsComputer = vsComputer
		a.score = [3]int{}
	}
}

func (a *App) updatePlaying(clicked bool) {
	select {
	case reply := <-a.aiReplies:
		if reply.match == a.match && a.thinking {
			a.thinking = false
			a.lastSearch = reply.result
			a.play(reply.result.Move)
		}
		return
	default:
	}

	if a.computerToPlay() {
		if !a.thinking {
			a.think()
		}
		return
	}
	if m, ok := moveAt(a.pointerX, a.pointerY); clicked && ok && a.game.Legal(m) {
		a.play(m)
	}
}

func (a *App) updateOver(clicked bool) {
	if !a.celebrated && a.clock >= a.endedAt+outcomeDelay {
		a.celebrated = true
		if a.outcome() == outcomeWin {
			a.confetti.burst(viewWidth/2, boardY+boardSize*0.45, 140)
		}
	}
	if a.clock < a.endedAt+cardDelay {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || (clicked && resultCardButton().contains(a.pointerX, a.pointerY)) {
		a.startGame()
	}
}

func (a *App) updateCursor() {
	shape := ebiten.CursorShapeDefault
	if a.hoveredMove() != engine.NoMove || a.hoveredButton() {
		shape = ebiten.CursorShapePointer
	}
	ebiten.SetCursorShape(shape)
}

// hoveredMove returns the legal move under the pointer when the player may play it.
func (a *App) hoveredMove() engine.Move {
	if a.phase != phasePlaying || a.computerToPlay() {
		return engine.NoMove
	}
	if m, ok := moveAt(a.pointerX, a.pointerY); ok && a.game.Legal(m) {
		return m
	}
	return engine.NoMove
}

func (a *App) hoveredButton() bool {
	x, y := a.pointerX, a.pointerY
	switch a.phase {
	case phaseMenu:
		menu := newMenuLayout()
		if menu.start.contains(x, y) || menu.opponents[0].contains(x, y) || menu.opponents[1].contains(x, y) {
			return true
		}
		for _, r := range menu.levels {
			if a.vsComputer && r.contains(x, y) {
				return true
			}
		}
		return false
	case phaseOver:
		if a.clock >= a.endedAt+cardDelay && resultCardButton().contains(x, y) {
			return true
		}
	}
	return newGameButton.contains(x, y)
}

func (a *App) computerToPlay() bool {
	return a.vsComputer && a.game.Turn() == computer
}

// think starts a search in the background; the reply is picked up by updatePlaying.
func (a *App) think() {
	a.thinking = true
	state, match, budget := a.game, a.match, time.Duration(a.level)*time.Second
	go func() {
		a.aiReplies <- aiReply{match: match, result: ai.Search(state, budget)}
	}()
}

func (a *App) play(m engine.Move) {
	a.game.Play(m)
	a.placedAt[m] = a.clock
	if a.game.BoardDone(m.Board()) {
		a.decidedAt[m.Board()] = a.clock
	}
	if a.game.Over() {
		a.phase = phaseOver
		a.endedAt = a.clock
		a.score[a.game.Winner()]++
	}
}

func (a *App) startGame() {
	a.match++
	a.game = engine.New(a.nextFirst)
	a.nextFirst = a.nextFirst.Opponent()
	a.thinking = false
	a.celebrated = false
	a.phase = phasePlaying
}

func (a *App) openMenu() {
	a.match++
	a.thinking = false
	a.phase = phaseMenu
}

func (a *App) outcome() outcome {
	switch w := a.game.Winner(); {
	case w == engine.Nobody:
		return outcomeDraw
	case a.vsComputer && w == computer:
		return outcomeLoss
	}
	return outcomeWin
}
