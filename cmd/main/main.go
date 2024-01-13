package main

import (
	"GoTicTacToe/lib/graphics"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	WindowWidth    = 800
	WindowHeight   = 900
	FontSize       = 15
	BigFontSize    = 100
	DPI            = 72
	NbPlayer       = 2
	BoardRowLength = 3
)

const (
	Init GameState = iota
	Playing
	PlayAgain
	WaitingForGameStart
)

// enum determining the symbols contained in the game
const (
	PLAYER1 GameSymbol = 'O'
	PLAYER2 GameSymbol = 'X'
	EMPTY   GameSymbol = ' ' // for empty cell
	NONE    GameSymbol = 0
)

var (
	normalText   font.Face
	bigText      font.Face
	symbolImage  *ebiten.Image
	gameImage    = ebiten.NewImage(WindowWidth, WindowWidth)
	gameGraphics = graphics.Init(WindowWidth)
)

// get the symbol of the cell at the given coordinates
func (g *Game) getValueOfCoordinates(coordinates graphics.BoardCoord) GameSymbol {
	return g.gameBoard[coordinates.MainBoardRow][coordinates.MainBoardCol].Board[coordinates.MiniBoardRow][coordinates.MiniBoardCol]
}

// set the symbol of the cell at the given coordinates
func (g *Game) setValueOfCoordinates(coordinates graphics.BoardCoord, value GameSymbol) {
	g.gameBoard[coordinates.MainBoardRow][coordinates.MainBoardCol].
		Board[coordinates.MiniBoardRow][coordinates.MiniBoardCol] = value
}

// get the coordinates of the cell clicked by the player in a mini tic-tac-toe board
func (g *Game) getMiniBoardCoordinates(mouseX, mouseY int) graphics.BoardCoord {
	miniTicTacToeSize := WindowWidth / BoardRowLength           // size of a whole mini tic-tac-toe board
	miniTicTacToeCellSize := miniTicTacToeSize / BoardRowLength // size of a cell in a mini tic-tac-toe board

	// get clicked cell coordinates
	// MAIN BOARD
	mainRow := mouseX / miniTicTacToeSize // the index of the row clicked
	mainCol := mouseY / miniTicTacToeSize // the index of the column clicked

	// get normalized coordinates
	normalizedX := mouseX - mainRow*miniTicTacToeSize
	normalizedY := mouseY - mainCol*miniTicTacToeSize

	// MINI BOARD
	miniRow := normalizedX / miniTicTacToeCellSize // the index of the row clicked
	miniCol := normalizedY / miniTicTacToeCellSize // the index of the column clicked

	return graphics.BoardCoord{
		MainBoardRow: mainRow,
		MainBoardCol: mainCol,
		MiniBoardRow: miniRow,
		MiniBoardCol: miniCol,
	}
}

// Update : game life cycle method called at "game tic" and apply the game logic depending on the current state.
// It is called by the ebiten engine.
func (g *Game) Update() error {
	switch g.state {
	case Init:
		// called at the beginning of the game
		g.init()
	case WaitingForGameStart:
		// At this point, the player is configuring the game parameters
		// before starting the game by clicking on the space bar
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = Playing
		}
		for i := ebiten.Key1; i <= ebiten.Key5; i++ {
			if inpututil.IsKeyJustPressed(i) {
				g.AIDifficulty = float64(i - ebiten.Key1 + 1)
				break
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			g.AIEnabled = !g.AIEnabled
		}
	case Playing:
		// At this point, the game is running and a player can make a move

		// if it is the AI's turn, we wait for it to finish
		// the AI is running in a goroutine
		if g.AIRunning {
			return nil
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			if mx > WindowWidth || my > WindowWidth {
				return nil
			}
			boardCoordinates := g.getMiniBoardCoordinates(mx, my)

			if !g.isValidPlay(boardCoordinates.MainBoardRow, boardCoordinates.MainBoardCol) {
				return nil
			}
			if g.getValueOfCoordinates(boardCoordinates) == EMPTY {
				g.makePlay(boardCoordinates)
			}
		}
		if g.AIEnabled && g.playing == PLAYER2 {
			go func() {
				g.AIRunning = true
				bestMove, simulations, winProbability := g.MonteCarloMove()
				g.AISimulations = simulations
				g.AIWinProbability = winProbability
				g.makePlay(bestMove)
				g.AIRunning = false
			}()
		}

	case PlayAgain:
		// At the end of a game, the player can choose to play again (clicking by mouse)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.Load()
		}
	}
	// at any time, the player can reset the game by pressing the R key or quit the game by pressing the escape key

	if inpututil.KeyPressDuration(ebiten.KeyR) == 60 {
		g.Load()
		g.ResetPoints()
	}
	if inpututil.KeyPressDuration(ebiten.KeyEscape) == 60 {
		os.Exit(0)
	}
	return nil
}

// determine if the move is applicable
func (g *Game) isValidPlay(row, col int) bool {
	if g.lastPlay.MiniBoardRow == -1 {
		// the first move of the game is always valid
		return true
	} else if g.gameBoard[g.lastPlay.MiniBoardRow][g.lastPlay.MiniBoardCol].Winner != EMPTY {
		// when the last move complete a mini-game, the next move can be played anywhere
		return true
	} else if row == g.lastPlay.MiniBoardRow && col == g.lastPlay.MiniBoardCol {
		// the next move must be played in the mini-game corresponding to the last move position
		return true
	}
	return false
}

func (g *Game) DrawSymbol(boardCoord graphics.BoardCoord, symbol GameSymbol) {
	symbolImage = g.getSymbolImage(symbol)

	xPos, yPos := graphics.GetPositionOfSymbol(boardCoord)
	opSymbol := &ebiten.DrawImageOptions{}
	opSymbol.GeoM.Translate(xPos, yPos)
	if g.lastPlay == boardCoord {
		opSymbol.ColorScale.Scale(0.5, 0.5, 0.5, 1)
	}
	gameImage.DrawImage(symbolImage, opSymbol)

}

func (g *Game) init() {
	// init font
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	normalText, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    FontSize,
		DPI:     DPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	bigText, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    BigFontSize,
		DPI:     DPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	re := newRandom().Intn(NbPlayer)
	if re == 0 {
		g.playing = PLAYER1
	} else {
		g.playing = PLAYER2
	}
	g.Load()
	g.ResetPoints()
	g.state = WaitingForGameStart
	g.lastPlay = graphics.BoardCoord{MainBoardRow: -1, MainBoardCol: -1, MiniBoardRow: -1, MiniBoardCol: -1}
	g.AIEnabled = true
}

func (g *Game) Load() {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			g.gameBoard[i][j] = MiniBoard{Board: [3][3]GameSymbol{
				{EMPTY, EMPTY, EMPTY},
				{EMPTY, EMPTY, EMPTY},
				{EMPTY, EMPTY, EMPTY}},
				Winner: EMPTY}
		}
	}
	g.round = 0
	g.win = EMPTY
	g.lastPlay = graphics.BoardCoord{MainBoardRow: -1, MainBoardCol: -1, MiniBoardRow: -1, MiniBoardCol: -1}

	// by default, the AI is set to the second difficulty level
	if g.AIDifficulty == 0 {
		g.AIDifficulty = 2
	}
	g.state = WaitingForGameStart
}

func (g *Game) wins(winner GameSymbol) {
	if winner == PLAYER1 {
		g.win = PLAYER1
		g.pointsO++
		g.state = PlayAgain
	} else if winner == PLAYER2 {
		g.win = PLAYER2
		g.pointsX++
		g.state = PlayAgain
	} else if winner == NONE {
		g.win = NONE
		g.state = PlayAgain
	}
}

func (g *Game) CheckWin() GameSymbol {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			g.gameBoard[i][j].CheckWin()
		}
	}
	for i := 0; i < 3; i++ {
		if g.winnerOnLine(i, 0, 0, 1) != EMPTY {
			return g.winnerOnLine(i, 0, 0, 1)
		}
	}
	for i := 0; i < 3; i++ {
		if g.winnerOnLine(0, i, 1, 0) != EMPTY {
			return g.winnerOnLine(0, i, 1, 0)
		}
	}
	if g.winnerOnLine(0, 0, 1, 1) != EMPTY {
		return g.winnerOnLine(0, 0, 1, 1)
	}
	if g.winnerOnLine(0, 2, 1, -1) != EMPTY {
		return g.winnerOnLine(0, 2, 1, -1)
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.gameBoard[i][j].Winner == EMPTY {
				return EMPTY
			}
		}
	}

	return NONE
}

// winnerOnLine checks if there is a winner on the given line
// x, y: the starting point of the line
// dx, dy: delta applied to x and y to get the next point on the line
func (g *Game) winnerOnLine(x, y, dx, dy int) GameSymbol {
	for i := 0; i < 3; i++ {
		if g.gameBoard[x][y].Winner != g.gameBoard[x+dx*i][y+dy*i].Winner {
			return EMPTY
		}
	}
	return g.gameBoard[x][y].Winner
}

func (g *Game) ResetPoints() {
	g.pointsO = 0
	g.pointsX = 0
}

func newRandom() *rand.Rand {
	s1 := rand.NewSource(time.Now().UnixNano())
	return rand.New(s1)
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WindowWidth, WindowHeight
}
func main() {
	game := &Game{}
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("TicTacToe")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) switchPlayer() {
	g.playing = g.getOpponents(g.playing)
}
func (g *Game) getOpponents(playerJustMoved GameSymbol) GameSymbol {
	if playerJustMoved == PLAYER1 {
		return PLAYER2
	}
	return PLAYER1
}
func (g *Game) makePlay(move graphics.BoardCoord) {
	g.setValueOfCoordinates(move, g.playing)
	g.wins(g.CheckWin())
	g.round++
	g.lastPlay = move
	g.switchPlayer()
}
func (g *Game) getSymbolImage(player GameSymbol) *ebiten.Image {
	if player == PLAYER1 {
		return gameGraphics.Circle
	}
	return gameGraphics.Cross
}
