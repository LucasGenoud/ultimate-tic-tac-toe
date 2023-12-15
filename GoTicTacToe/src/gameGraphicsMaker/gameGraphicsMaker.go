package gameGraphicsMaker

import (
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	boardSize       = 480
	symbolSize      = boardSize / 3
	symbolLineWidth = 10
	boardLineWidth  = 5
)

type GameGraphics struct {
	Board  *ebiten.Image
	Circle *ebiten.Image
	Cross  *ebiten.Image
}

func Init() GameGraphics {
	symbols := GameGraphics{}
	symbols.Circle = drawCircle()
	symbols.Cross = drawCross()
	symbols.Board = DrawBoard()

	return symbols
}

func DrawBoard() *ebiten.Image {
	context := gg.NewContext(boardSize, boardSize)
	rectangleBaseSize := float64(boardSize / 3)
	context.SetRGBA(1, 1, 1, 1)
	context.DrawRectangle(rectangleBaseSize-boardLineWidth, 0, boardLineWidth, boardSize)
	context.DrawRectangle(rectangleBaseSize*2-boardLineWidth, 0, boardLineWidth, boardSize)
	context.DrawRectangle(0, rectangleBaseSize-boardLineWidth, boardSize, boardLineWidth)
	context.DrawRectangle(0, rectangleBaseSize*2-boardLineWidth, boardSize, boardLineWidth)
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
	context.DrawRectangle(0, symbolSize/2-symbolLineWidth/2, symbolSize, symbolLineWidth)
	context.DrawRectangle(symbolSize/2-symbolLineWidth/2, 0, symbolLineWidth, symbolSize)
	context.Fill()

	return ebiten.NewImageFromImage(context.Image())
}
func GetPositionOfSymbol(x, y int) (float64, float64) {
	return float64(symbolSize*x - boardLineWidth*x/2), float64(symbolSize * y)
}
