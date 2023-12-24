package main

import (
	"GoTicTacToe/lib/graphics"
	"GoTicTacToe/lib/models"
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
	sWidth      = 800
	sHeight     = 900
	fontSize    = 15
	bigFontSize = 100
	dpi         = 72
	nbPlayer    = 2
)

type GameState int

const (
	Init GameState = iota
	Playing
	PlayAgain
)

// enum determining the symbols contained in the game
const (
	PLAYER1 models.GameSymbol = 'O'
	PLAYER2 models.GameSymbol = 'X'
	EMPTY   models.GameSymbol = ' ' // for empty cell
	NONE    models.GameSymbol = 0
)

var (
	normalText   font.Face
	bigText      font.Face
	boardImage   *ebiten.Image
	symbolImage  *ebiten.Image
	gameImage    = ebiten.NewImage(sWidth, sWidth)
	gameGraphics = graphics.Init(sWidth)
)

type Game struct {
	playing   models.GameSymbol
	state     GameState
	gameBoard [3][3]MiniBoard
	round     int
	pointsO   int
	pointsX   int
	win       models.GameSymbol
	lastPlay  graphics.BoardCoord
}

func (g *Game) getValueOfCoordinates(coordinates graphics.BoardCoord) models.GameSymbol {
	return g.gameBoard[coordinates.MainBoardRow][coordinates.MainBoardCol].Board[coordinates.MiniBoardRow][coordinates.MiniBoardCol]
}
func (g *Game) setValueOfCoordinates(coordinates graphics.BoardCoord, value models.GameSymbol) {
	g.gameBoard[coordinates.MainBoardRow][coordinates.MainBoardCol].Board[coordinates.MiniBoardRow][coordinates.MiniBoardCol] = value
}
func (g *Game) getMiniBoardCoordinates(mouseX, mouseY int) graphics.BoardCoord {
	miniTicTacToeSize := sWidth / 3
	ticTacToeCellSize := miniTicTacToeSize / 3
	mainRow := mouseX / miniTicTacToeSize
	mainCol := mouseY / miniTicTacToeSize
	normalizedX := mouseX - mainRow*miniTicTacToeSize
	normalizedY := mouseY - mainCol*miniTicTacToeSize
	miniRow := normalizedX / ticTacToeCellSize
	miniCol := normalizedY / ticTacToeCellSize
	return graphics.BoardCoord{MainBoardRow: mainRow, MainBoardCol: mainCol, MiniBoardRow: miniRow, MiniBoardCol: miniCol}

}

func (g *Game) Update() error {
	switch g.state {
	case Init:
		g.init()
	case Playing:

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			if mx > sWidth || my > sWidth {
				return nil
			}
			boardCoordinates := g.getMiniBoardCoordinates(mx, my)

			if !g.isValidPlay(boardCoordinates.MainBoardRow, boardCoordinates.MainBoardCol) {
				return nil
			}
			if g.getValueOfCoordinates(boardCoordinates) == EMPTY {
				g.lastPlay = boardCoordinates
				if g.playing == PLAYER1 {
					g.setValueOfCoordinates(boardCoordinates, PLAYER1)
					g.playing = PLAYER2
				} else {
					g.setValueOfCoordinates(boardCoordinates, PLAYER2)
					g.playing = PLAYER1
				}
				g.wins(g.CheckWin())
				g.round++
			}
		}

	case PlayAgain:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.Load()
		}
	}
	if inpututil.KeyPressDuration(ebiten.KeyR) == 60 {
		g.Load()
		g.ResetPoints()
	}
	if inpututil.KeyPressDuration(ebiten.KeyEscape) == 60 {
		os.Exit(0)
	}
	return nil
}

func (g *Game) isValidPlay(row, col int) bool {
	if g.lastPlay.MiniBoardRow == -1 {
		return true
	} else if g.gameBoard[g.lastPlay.MiniBoardRow][g.lastPlay.MiniBoardCol].Winner != EMPTY {
		return true
	} else if row == g.lastPlay.MiniBoardRow && col == g.lastPlay.MiniBoardCol {
		return true
	}
	return false
}

func (g *Game) DrawSymbol(boardCoord graphics.BoardCoord, symbol models.GameSymbol) {
	if symbol == PLAYER1 {
		symbolImage = gameGraphics.Circle
	}
	if symbol == PLAYER2 {
		symbolImage = gameGraphics.Cross
	}

	xPos, yPos := graphics.GetPositionOfSymbol(boardCoord)
	opSymbol := &ebiten.DrawImageOptions{}
	opSymbol.GeoM.Translate(xPos, yPos)

	gameImage.DrawImage(symbolImage, opSymbol)

}

func (g *Game) init() {
	// init font
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}
	normalText, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	bigText, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    bigFontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	// init game state
	boardImage = gameGraphics.MainBoard

	re := newRandom().Intn(nbPlayer)
	if re == 0 {
		g.playing = PLAYER1
	} else {
		g.playing = PLAYER2
	}
	g.Load()
	g.ResetPoints()
	g.lastPlay = graphics.BoardCoord{MainBoardRow: -1, MainBoardCol: -1, MiniBoardRow: -1, MiniBoardCol: -1}
}

func (g *Game) Load() {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			g.gameBoard[i][j] = MiniBoard{Board: [3][3]models.GameSymbol{{EMPTY, EMPTY, EMPTY}, {EMPTY, EMPTY, EMPTY}, {EMPTY, EMPTY, EMPTY}}, Winner: EMPTY}
		}
	}
	g.round = 0
	g.win = EMPTY
	g.state = Playing
	g.lastPlay = graphics.BoardCoord{MainBoardRow: -1, MainBoardCol: -1, MiniBoardRow: -1, MiniBoardCol: -1}

}

func (g *Game) wins(winner models.GameSymbol) {
	if winner == PLAYER1 {
		g.win = PLAYER1
		g.pointsO++
		g.state = PlayAgain
	} else if winner == PLAYER2 {
		g.win = PLAYER2
		g.pointsX++
		g.state = PlayAgain
	}
}

func (g *Game) CheckWin() models.GameSymbol {
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
	return NONE
}

// winnerOnLine checks if there is a winner on the given line
// x, y: the starting point of the line
// dx, dy: delta applied to x and y to get the next point on the line
func (g *Game) winnerOnLine(x, y, dx, dy int) models.GameSymbol {
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
	return sWidth, sHeight
}
func main() {
	game := &Game{}
	ebiten.SetWindowSize(sWidth, sHeight)
	ebiten.SetWindowTitle("TicTacToe")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
