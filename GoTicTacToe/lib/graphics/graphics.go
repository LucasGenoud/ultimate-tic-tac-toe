package graphics

import (
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	boardSize       = 480
	numberOfRows    = 3
	symbolSize      = boardSize / numberOfRows
	symbolLineWidth = 10
	boardLineWidth  = 10
)

type GameGraphics struct {
	Board  *ebiten.Image
	Circle *ebiten.Image
	Cross  *ebiten.Image
}

func Init() GameGraphics {
	gameGraphics := GameGraphics{}
	gameGraphics.Circle = drawCircle()
	gameGraphics.Cross = drawCross()
	gameGraphics.Board = DrawBoard()

	return gameGraphics
}

func DrawBoard() *ebiten.Image {
	context := gg.NewContext(boardSize, boardSize)
	boardCaseSize := float64(boardSize / numberOfRows)
	context.SetRGBA(1, 1, 1, 1)

	for i := 1; i < numberOfRows; i++ {
		context.DrawRectangle(boardCaseSize*float64(i)-boardLineWidth/2, 0, boardLineWidth, boardSize)
		context.DrawRectangle(0, boardCaseSize*float64(i)-boardLineWidth/2, boardSize, boardLineWidth)

	}

	context.Fill()
	return ebiten.NewImageFromImage(context.Image())
}

func drawCircle() *ebiten.Image {
	const radius = symbolSize/2 - boardLineWidth*2

	context := gg.NewContext(symbolSize, symbolSize)
	context.SetRGBA(1, 1, 1, 1)
	context.SetLineWidth(symbolLineWidth)

	context.DrawCircle(symbolSize/2, symbolSize/2, radius)
	context.Stroke()

	return ebiten.NewImageFromImage(context.Image())
}

func drawCross() *ebiten.Image {

	context := gg.NewContext(symbolSize, symbolSize)
	context.SetRGBA(1, 1, 1, 1)
	context.RotateAbout(gg.Radians(45), symbolSize/2, symbolSize/2)

	context.DrawRectangle(0, symbolSize/2-symbolLineWidth/2, symbolSize, symbolLineWidth)
	context.DrawRectangle(symbolSize/2-symbolLineWidth/2, 0, symbolLineWidth, symbolSize)
	context.Fill()

	return ebiten.NewImageFromImage(context.Image())
}

func GetPositionOfSymbol(x, y int) (float64, float64) {
	return float64(symbolSize*x - boardLineWidth*x/2), float64(symbolSize * y)
}
