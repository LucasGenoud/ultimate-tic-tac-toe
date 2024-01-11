package main

import (
	"GoTicTacToe/lib/graphics"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"image/color"
)

func (g *Game) Draw(screen *ebiten.Image) {
	gameImage.Clear()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.gameBoard[i][j].Winner == EMPTY {
				g.drawMiniBoard(i, j, screen)
			} else {
				g.drawMiniBoardWinner(i, j, screen)
			}
		}
	}
	gameImage.DrawImage(gameGraphics.MainBoard, nil)
	screen.DrawImage(gameImage, nil)

	g.displayInformation(screen)
	g.drawAIRunning(screen)
}
func (g *Game) drawMiniBoardWinner(i, j int, screen *ebiten.Image) {
	gameBoardImageOptions := &ebiten.DrawImageOptions{}

	gameBoardImageOptions.GeoM.Reset()
	gameBoardImageOptions.GeoM.Scale(3, 3)
	gameBoardImageOptions.GeoM.Translate(float64(sWidth/3*i), float64(sWidth/3*j))
	if g.gameBoard[i][j].Winner == PLAYER1 {
		screen.DrawImage(gameGraphics.Circle, gameBoardImageOptions)
	} else {
		screen.DrawImage(gameGraphics.Cross, gameBoardImageOptions)
	}
}
func (g *Game) drawMiniBoard(i, j int, screen *ebiten.Image) {

	for k := 0; k < 3; k++ {
		for l := 0; l < 3; l++ {
			symbolInCell := g.gameBoard[i][j].Board[k][l]
			if symbolInCell == PLAYER1 || symbolInCell == PLAYER2 {
				g.DrawSymbol(graphics.BoardCoord{MainBoardRow: i, MainBoardCol: j, MiniBoardRow: k, MiniBoardCol: l}, symbolInCell)
			}
		}
	}

	gameBoardImageOptions := &ebiten.DrawImageOptions{}
	gameBoardImageOptions.GeoM.Translate(float64(sWidth/3*i), float64(sWidth/3*j))
	if g.isValidPlay(i, j) {
		gameBoardImageOptions.ColorScale.Scale(0, 1, 0, 1)
	}

	screen.DrawImage(gameGraphics.MiniBoard, gameBoardImageOptions)
}
func (g *Game) displayInformation(screen *ebiten.Image) {
	mx, my := ebiten.CursorPosition()

	msgFPS := fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f", ebiten.ActualTPS(), ebiten.ActualFPS())
	text.Draw(screen, msgFPS, normalText, 0, sHeight-30, color.White)

	if g.AIEnabled {
		msgAI := fmt.Sprintf("AI simulations: %v \nAI win confidence: %0.2f\nAI difficulty: %v ", g.AISimulations, g.AIWinProbability*100, g.AIDifficulty)
		text.Draw(screen, msgAI, normalText, 100, sHeight-50, color.White)
	}

	keyChangeColor(ebiten.KeyEscape, screen)
	keyChangeColor(ebiten.KeyR, screen)
	msgOX := fmt.Sprintf("O: %v | X: %v", g.pointsO, g.pointsX)
	text.Draw(screen, msgOX, normalText, sWidth/2, sHeight-5, color.White)
	if g.win != EMPTY {
		var msgWin = ""
		if g.win == NONE {
			msgWin = "Draw!"
		} else {
			msgWin = fmt.Sprintf("%v wins!", string(g.win))
		}
		text.Draw(screen, msgWin, bigText, 70, 200, color.RGBA{G: 50, B: 200, A: 255})
	}
	if g.state == WaitingForGameStart {
		msg := ""
		if g.AIEnabled {
			msg = "Press SPACE to start\nPress A to switch to multiplayer\nPress 1 to 5 to change AI difficulty"

		} else {
			msg = "Press SPACE to start\nPress A to enable AI"

		}
		widthX, _ := font.BoundString(normalText, msg)

		text.Draw(screen, msg, normalText, int(sWidth/2-widthX.Min.X), sHeight/2, color.RGBA{0, 255, 255, 255})
	}
	msg := string(g.playing)
	text.Draw(screen, msg, normalText, mx, my, color.RGBA{G: 255, A: 255})
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
func (g *Game) drawAIRunning(screen *ebiten.Image) {
	if g.AIRunning {
		msg := "AI is running..."
		x := (sWidth - len(msg)*fontSize) / 2
		y := sHeight / 2
		text.Draw(screen, msg, normalText, x, y, color.White)
	}
}
