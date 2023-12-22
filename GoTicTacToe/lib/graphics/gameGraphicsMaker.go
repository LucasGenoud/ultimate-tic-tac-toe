package graphics

import (
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

type gameGraphicMaker struct {
	context *gg.Context
}

func (ggm *gameGraphicMaker) drawRectangle(x, y, width, height int) {
	ggm.context.DrawRectangle(float64(x), float64(y), float64(width), float64(height))
}
func (ggm *gameGraphicMaker) fill() {
	ggm.context.Fill()
}
func (ggm *gameGraphicMaker) setRGBA(r, g, b, a float64) {
	ggm.context.SetRGBA(r, g, b, a)
}
func (ggm *gameGraphicMaker) drawCircle(x, y, radius int) {
	ggm.context.DrawCircle(float64(x), float64(y), float64(radius))
}
func (ggm *gameGraphicMaker) stroke() {
	ggm.context.Stroke()
}
func (ggm *gameGraphicMaker) rotateAbout(angle, x, y int) {
	ggm.context.RotateAbout(gg.Radians(float64(angle)), float64(x), float64(y))
}
func (ggm *gameGraphicMaker) setLineWidth(width float64) {
	ggm.context.SetLineWidth(width)
}
func (ggm *gameGraphicMaker) getImage() *ebiten.Image {
	return ebiten.NewImageFromImage(ggm.context.Image())
}
