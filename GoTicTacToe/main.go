package main

import (
	"GoTicTacToe/src/gameGraphicsMaker"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image/color"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	sWidth      = 480
	sHeight     = 600
	fontSize    = 15
	bigFontSize = 100
	dpi         = 72
	nbPlayer    = 2
)

type GameState int

const (
	Init GameState = iota
	AITurn
	PlayerTurn
)

type PlayerSymbol rune

const (
	PLAYER1 PlayerSymbol = 'O'
	PLAYER2 PlayerSymbol = 'X'
	EMPTY   PlayerSymbol = ' '
	NONE    PlayerSymbol = 0
)

var (
	normalText   font.Face
	bigText      font.Face
	boardImage   *ebiten.Image
	symbolImage  *ebiten.Image
	gameImage    = ebiten.NewImage(sWidth, sWidth)
	gameGraphics = gameGraphicsMaker.Init()
)

type Game struct {
	playing   PlayerSymbol
	state     GameState
	gameBoard [3][3]PlayerSymbol
	round     int
	pointsO   int
	pointsX   int
	win       PlayerSymbol
	alter     int
}

func (g *Game) Update() error {
	switch g.state {
	case Init:
		g.Init()
	case AITurn:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			if mx/160 < 3 && mx >= 0 && my/160 < 3 && my >= 0 && g.gameBoard[mx/160][my/160] == EMPTY {
				if g.round%2 == 0+g.alter {
					g.DrawSymbol(mx/160, my/160, string(rune(PLAYER1)))
					g.gameBoard[mx/160][my/160] = PLAYER1
					g.playing = PLAYER2
				} else {
					g.DrawSymbol(mx/160, my/160, string(rune(PLAYER2)))
					g.gameBoard[mx/160][my/160] = PLAYER2
					g.playing = PLAYER1
				}
				g.wins(g.CheckWin())
				g.round++
			}
		}
	case PlayerTurn:
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

func keyChangeColor(key ebiten.Key, screen *ebiten.Image) {
	if inpututil.KeyPressDuration(key) > 1 {
		var msgText string
		var colorText color.RGBA
		colorChange := 255 - (255 / 60 * uint8(inpututil.KeyPressDuration(key)))
		if key == ebiten.KeyEscape {
			msgText = "CLOSING..."
			colorText = color.RGBA{R: 255, G: colorChange, B: colorChange, A: 255}
		} else if key == ebiten.KeyR {
			msgText = "RESETING..."
			colorText = color.RGBA{R: colorChange, G: 255, B: 255, A: 255}
		}
		text.Draw(screen, msgText, normalText, sWidth/2, sHeight-30, colorText)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {

	screen.DrawImage(boardImage, nil)
	screen.DrawImage(gameImage, nil)
	mx, my := ebiten.CursorPosition()

	msgFPS := fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f", ebiten.ActualTPS(), ebiten.ActualFPS())
	text.Draw(screen, msgFPS, normalText, 0, sHeight-30, color.White)

	keyChangeColor(ebiten.KeyEscape, screen)
	keyChangeColor(ebiten.KeyR, screen)
	msgOX := fmt.Sprintf("O: %v | X: %v", g.pointsO, g.pointsX)
	text.Draw(screen, msgOX, normalText, sWidth/2, sHeight-5, color.White)
	if g.win != EMPTY {
		msgWin := fmt.Sprintf("%v wins!", string(g.win))
		text.Draw(screen, msgWin, bigText, 70, 200, color.RGBA{G: 50, B: 200, A: 255})
	}
	msg := string(g.playing)
	text.Draw(screen, msg, normalText, mx, my, color.RGBA{G: 255, A: 255})
}

func (g *Game) DrawSymbol(x, y int, sym string) {
	//TODO : sym may be replace by rune or playerSymbol one images would be generate
	if sym == "X" {
		symbolImage = gameGraphics.Cross
	}
	if sym == "O" {
		symbolImage = gameGraphics.Circle
	}

	xPos, yPos := gameGraphicsMaker.GetPositionOfSymbol(x, y)
	opSymbol := &ebiten.DrawImageOptions{}
	opSymbol.GeoM.Translate(xPos, yPos)

	gameImage.DrawImage(symbolImage, opSymbol)
}

func (g *Game) Init() {
	boardImage = gameGraphics.Board
	re := newRandom().Intn(nbPlayer)
	if re == 0 {
		g.playing = PLAYER1
		g.alter = 0
	} else {
		g.playing = PLAYER2
		g.alter = 1
	}
	g.Load()
	g.ResetPoints()
}

func (g *Game) Load() {
	gameImage.Clear()
	g.gameBoard = [3][3]PlayerSymbol{{EMPTY, EMPTY, EMPTY}, {EMPTY, EMPTY, EMPTY}, {EMPTY, EMPTY, EMPTY}}
	g.round = 0
	if g.alter == 0 {
		g.playing = PLAYER2
		g.alter = 1
	} else if g.alter == 1 {
		g.playing = PLAYER1
		g.alter = 0
	}
	g.win = EMPTY
	g.state = AITurn
}

func (g *Game) wins(winner PlayerSymbol) {
	if winner == PLAYER1 {
		g.win = PLAYER1
		g.pointsO++
		g.state = PlayerTurn
	} else if winner == PLAYER2 {
		g.win = PLAYER2
		g.pointsX++
		g.state = PlayerTurn
	} else if winner == NONE {
		g.win = NONE
		g.state = PlayerTurn
	}
}

func (g *Game) CheckWin() PlayerSymbol {

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
	if g.round == 8 {
		return NONE
	}
	return EMPTY
}

// winnerOnLine checks if there is a winner on the given line
// x, y: the starting point of the line
// dx, dy: delta applied to x and y to get the next point on the line
func (g *Game) winnerOnLine(x, y, dx, dy int) PlayerSymbol {
	for i := 0; i < 3; i++ {
		if g.gameBoard[x][y] != g.gameBoard[x+dx*i][y+dy*i] {
			return EMPTY
		}
	}
	return g.gameBoard[x][y]
}

func (g *Game) ResetPoints() {
	g.pointsO = 0
	g.pointsX = 0
}

func init() {
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
}

func newRandom() *rand.Rand {
	s1 := rand.NewSource(time.Now().UnixNano())
	return rand.New(s1)
}

func (g *Game) Layout(int, int) (int, int) {
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
