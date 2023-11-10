package symbolDrawer

import (
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

type Symbols struct {
	Board  *ebiten.Image
	Circle *ebiten.Image
	Cross  *ebiten.Image
}

func Init() Symbols {
	symbols := Symbols{}
	symbols.Circle = drawCircle()
	symbols.Board = DrawBoard()

	return symbols
}
func DrawBoard() *ebiten.Image {
	const S = 1024
	context := gg.NewContext(S, S)
	context.SetRGBA(10, 10, 10, 0.1)

	context.DrawRectangle(1024/3, 0, 10, S)
	context.DrawRectangle(1024/3*2, 0, 10, S)
	return ebiten.NewImageFromImage(context.Image())
}

func drawCircle() *ebiten.Image {
	const S = 1024
	context := gg.NewContext(S, S)
	context.SetRGBA(10, 10, 10, 0.1)
	for i := 0; i < 360; i += 5 {
		context.Push()
		context.RotateAbout(gg.Radians(float64(i)), S/2, S/2)
		context.DrawEllipse(S/2, S/2, S*7/16, S/8)
		context.Fill()
		context.Pop()
	}
	return ebiten.NewImageFromImage(context.Image())
}
