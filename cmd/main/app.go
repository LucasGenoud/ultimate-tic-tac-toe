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
	human    = engine.O
	computer = engine.X
	// minThinkTime keeps the computer from answering instantly at the low levels, in seconds.
	minThinkTime = 0.6
	// progressPollTicks is how often the statistics of a running search are refreshed.
	progressPollTicks = 6
)

var hintLimits = ai.Limits{Duration: 1500 * time.Millisecond}

type requestKind int

const (
	requestMove requestKind = iota // the computer's move
	requestHint                    // a suggestion for the player
)

// request is a search running in the background.
type request struct {
	kind    requestKind
	limits  ai.Limits
	started float64
	result  *ai.Result // set once the search has returned
}

type aiReply struct {
	match  int
	result ai.Result
}

// App is the ebiten game: it owns the match, the menus and the animations.
type App struct {
	settings  settings
	game      engine.Game
	history   []engine.Move
	first     engine.Player // who started the current game
	nextFirst engine.Player // who starts the next game when alternating
	phase     phase
	match     int // incremented whenever searches are cancelled, so their replies are ignored

	searcher   *ai.Searcher
	replies    chan aiReply
	request    *request
	pondering  bool
	live       ai.Result // statistics of the running search
	lastSearch ai.Result // statistics of the computer's last move
	hint       engine.Move
	hintResult ai.Result

	sounds         *sounds
	showRules      bool
	draggingSlider bool

	ticks      int
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
	s := loadSettings()
	first := engine.O
	if rand.IntN(2) == 0 {
		first = engine.X
	}
	a := &App{
		settings:  s,
		game:      engine.New(first),
		first:     first,
		nextFirst: first,
		searcher:  ai.NewSearcher(),
		replies:   make(chan aiReply, 4),
		hint:      engine.NoMove,
		sounds:    newSounds(),
		showRules: !s.SeenRules,
	}
	a.sounds.mute = !s.Sound
	return a
}

func (a *App) pal() *palette {
	if a.settings.Dark {
		return &darkPalette
	}
	return &lightPalette
}

func (a *App) level() level { return levels[a.settings.Level] }

// threads returns how many threads the computer uses.
func (a *App) threads() int {
	if t := a.settings.Threads; t > 0 && t < a.searcher.MaxThreads() {
		return t
	}
	return a.searcher.MaxThreads()
}

func (a *App) save() { writeSettings(a.settings) }

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
	a.ticks++
	a.clock += dt
	a.confetti.update(dt)
	a.pointerX, a.pointerY = a.viewport.ToVirtual(ebiten.CursorPositionF())
	if touches := ebiten.AppendTouchIDs(nil); len(touches) > 0 {
		a.pointerX, a.pointerY = a.viewport.ToVirtual(ebiten.TouchPositionF(touches[0]))
	}
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || len(inpututil.AppendJustPressedTouchIDs(nil)) > 0

	if err := a.handleKeys(); err != nil {
		return err
	}
	a.receiveReplies()

	switch {
	case a.showRules:
		a.updateRules(clicked)
	case a.phase == phaseMenu:
		a.updateMenu(clicked)
	case clicked && a.clickFooter():
	case a.phase == phasePlaying:
		a.updatePlaying(clicked)
	case a.phase == phaseOver:
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

func (a *App) handleKeys() error {
	hold := ebiten.TPS() // one second
	if inpututil.KeyPressDuration(ebiten.KeyR) == hold {
		a.settings.Score = [3]int{}
		a.save()
	}
	if runtime.GOOS != "js" && inpututil.KeyPressDuration(ebiten.KeyEscape) == hold {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		a.setSound(!a.settings.Sound)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		a.settings.Dark = !a.settings.Dark
		a.save()
	}
	if a.showRules || a.phase == phaseMenu {
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyH) && a.canHint() {
		a.startRequest(requestHint)
	}
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if inpututil.IsKeyJustPressed(ebiten.KeyU) || (ctrl && inpututil.IsKeyJustPressed(ebiten.KeyZ)) {
		a.undo()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		a.openMenu()
	}
	return nil
}

// receiveReplies hands finished searches to the current request, dropping cancelled ones.
func (a *App) receiveReplies() {
	for {
		select {
		case reply := <-a.replies:
			if a.request != nil && reply.match == a.match {
				a.request.result = &reply.result
			}
		default:
			return
		}
	}
}

func (a *App) clickFooter() bool {
	x, y := a.pointerX, a.pointerY
	switch {
	case hintButton.contains(x, y):
		if a.canHint() {
			a.startRequest(requestHint)
		}
	case undoButton.contains(x, y):
		a.undo()
	case newGameButton.contains(x, y):
		a.openMenu()
	default:
		return false
	}
	return true
}

func (a *App) updatePlaying(clicked bool) {
	a.updateRequest()
	if a.phase != phasePlaying {
		return
	}
	if a.computerToPlay() {
		if a.request == nil {
			a.startRequest(requestMove)
		}
		return
	}
	if a.request == nil && !a.pondering && a.settings.VsComputer && a.level().ponder {
		a.searcher.Ponder(a.game, a.threads())
		a.pondering = true
	}
	if m := a.hoveredMove(); clicked && m != engine.NoMove {
		if a.request != nil {
			a.cancelSearches() // a hint was on its way
		}
		a.play(m)
	}
}

// updateRequest refreshes the statistics of the running search and applies its result.
func (a *App) updateRequest() {
	r := a.request
	if r == nil {
		return
	}
	if r.result == nil {
		if r.kind == requestMove && a.ticks%progressPollTicks == 0 {
			a.live = a.searcher.Progress()
		}
		return
	}
	switch r.kind {
	case requestMove:
		a.live = *r.result
		if a.clock < r.started+minThinkTime {
			return
		}
		a.request = nil
		a.lastSearch = *r.result
		a.play(r.result.Move)
	case requestHint:
		a.request = nil
		a.hint, a.hintResult = r.result.Move, *r.result
	}
}

func (a *App) updateOver(clicked bool) {
	if !a.celebrated && a.clock >= a.endedAt+outcomeDelay {
		a.celebrated = true
		switch a.outcome() {
		case outcomeWin:
			a.confetti.burst(viewWidth/2, boardY+boardSize*0.45, 140, a.pal().confetti)
			a.sounds.play(soundWin)
		case outcomeLoss:
			a.sounds.play(soundLoss)
		case outcomeDraw:
			a.sounds.play(soundDraw)
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
	if a.hoveredMove() != engine.NoMove || a.hoveredControl() {
		shape = ebiten.CursorShapePointer
	}
	ebiten.SetCursorShape(shape)
}

// hoveredMove returns the legal move under the pointer when the player may play it.
func (a *App) hoveredMove() engine.Move {
	if a.phase != phasePlaying || a.showRules || a.computerToPlay() {
		return engine.NoMove
	}
	if m, ok := moveAt(a.pointerX, a.pointerY); ok && a.game.Legal(m) {
		return m
	}
	return engine.NoMove
}

func (a *App) hoveredControl() bool {
	x, y := a.pointerX, a.pointerY
	switch {
	case a.showRules:
		return rulesButton().contains(x, y)
	case a.phase == phaseMenu:
		return a.menuControlAt(x, y)
	case a.phase == phaseOver && a.clock >= a.endedAt+cardDelay && resultCardButton().contains(x, y):
		return true
	}
	return (hintButton.contains(x, y) && a.canHint()) || (undoButton.contains(x, y) && a.canUndo()) || newGameButton.contains(x, y)
}

func (a *App) computerToPlay() bool {
	return a.settings.VsComputer && a.game.Turn() == computer
}

// startRequest searches the current position in the background; updateRequest picks up the result.
func (a *App) startRequest(kind requestKind) {
	a.pondering = false
	limits := hintLimits
	if kind == requestMove {
		limits = a.level().limits
		a.live = ai.Result{}
	}
	limits.Threads = a.threads()
	a.request = &request{kind: kind, limits: limits, started: a.clock}
	state, match := a.game, a.match
	go func() {
		a.replies <- aiReply{match: match, result: a.searcher.Search(state, limits)}
	}()
}

// cancelSearches stops the running search or pondering and ignores its reply.
func (a *App) cancelSearches() {
	a.match++
	a.request = nil
	a.pondering = false
	a.searcher.Cancel()
}

func (a *App) canHint() bool {
	return a.phase == phasePlaying && !a.computerToPlay() && a.request == nil
}

func (a *App) play(m engine.Move) {
	mover := a.game.Turn()
	a.game.Play(m)
	a.history = append(a.history, m)
	a.hint = engine.NoMove
	a.placedAt[m] = a.clock

	board := m.Board()
	switch {
	case a.game.BoardWinner(board) != engine.Nobody:
		a.decidedAt[board] = a.clock
		a.sounds.play(soundBoardWon)
	case a.game.BoardDone(board):
		a.decidedAt[board] = a.clock
		a.sounds.play(soundBoardDrawn)
	case mover == engine.O:
		a.sounds.play(soundO)
	default:
		a.sounds.play(soundX)
	}

	if a.game.Over() {
		a.cancelSearches()
		a.phase = phaseOver
		a.endedAt = a.clock
		a.settings.Score[a.game.Winner()]++
		a.save()
	}
}

// undoTarget returns how many moves of the history to keep when undoing: the player's last move,
// and the computer's answer to it.
func (a *App) undoTarget() (int, bool) {
	if a.phase == phaseMenu || len(a.history) == 0 {
		return 0, false
	}
	keep := len(a.history) - 1
	if !a.settings.VsComputer {
		return keep, true
	}
	for ; keep >= 0; keep-- {
		if g := a.replay(keep); g.Turn() == human {
			return keep, true
		}
	}
	return 0, false
}

func (a *App) canUndo() bool {
	_, ok := a.undoTarget()
	return ok
}

func (a *App) undo() {
	keep, ok := a.undoTarget()
	if !ok {
		return
	}
	a.cancelSearches()
	if a.phase == phaseOver {
		a.settings.Score[a.game.Winner()]--
		a.save()
	}
	a.history = a.history[:keep]
	a.game = a.replay(keep)
	a.phase = phasePlaying
	a.celebrated = false
	a.hint = engine.NoMove
	a.lastSearch = ai.Result{}
}

// replay returns the game after the first n moves of the history.
func (a *App) replay(n int) engine.Game {
	g := engine.New(a.first)
	for _, m := range a.history[:n] {
		g.Play(m)
	}
	return g
}

func (a *App) startGame() {
	a.cancelSearches()
	switch a.settings.FirstMove {
	case firstO:
		a.first = engine.O
	case firstX:
		a.first = engine.X
	default:
		a.first, a.nextFirst = a.nextFirst, a.nextFirst.Opponent()
	}
	a.game = engine.New(a.first)
	a.history = a.history[:0]
	a.hint = engine.NoMove
	a.live, a.lastSearch = ai.Result{}, ai.Result{}
	a.celebrated = false
	a.phase = phasePlaying
}

func (a *App) openMenu() {
	a.cancelSearches()
	a.phase = phaseMenu
}

func (a *App) setSound(on bool) {
	a.settings.Sound = on
	a.sounds.mute = !on
	a.save()
}

// setOpponent switches between playing the computer and a friend, starting a new score.
func (a *App) setOpponent(vsComputer bool) {
	if a.settings.VsComputer != vsComputer {
		a.settings.VsComputer = vsComputer
		a.settings.Score = [3]int{}
	}
}

func (a *App) outcome() outcome {
	switch w := a.game.Winner(); {
	case w == engine.Nobody:
		return outcomeDraw
	case a.settings.VsComputer && w == computer:
		return outcomeLoss
	}
	return outcomeWin
}
