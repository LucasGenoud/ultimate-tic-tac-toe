package main

import (
	"image/color"

	"GoTicTacToe/lib/engine"
)

// Everything is laid out in a virtual 800 x 1000 space, scaled to fit the window.
const (
	viewWidth  = 800.0
	viewHeight = 1000.0

	boardX       = 40.0
	boardY       = 140.0
	boardSize    = 720.0
	boardPadding = 20.0
	gridX        = boardX + boardPadding
	gridY        = boardY + boardPadding
	gridSize     = boardSize - 2*boardPadding
	bigCell      = gridSize / 3
	miniInset    = 14.0
	miniSize     = bigCell - 2*miniInset
	smallCell    = miniSize / 3

	headerY = 72.0
	footerY = 920.0
)

var (
	colorBackground = color.NRGBA{0xf4, 0xf2, 0xee, 0xff}
	colorSurface    = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	colorBorder     = color.NRGBA{0xe2, 0xdf, 0xd8, 0xff}
	colorInk        = color.NRGBA{0x24, 0x26, 0x2b, 0xff}
	colorMuted      = color.NRGBA{0x6e, 0x71, 0x7a, 0xff}
	colorGridMajor  = color.NRGBA{0x3a, 0x3d, 0x44, 0xff}
	colorGridMinor  = color.NRGBA{0xdc, 0xd8, 0xd0, 0xff}
	colorShadow     = color.NRGBA{0x00, 0x00, 0x00, 0x14}
	colorO          = color.NRGBA{0xcf, 0x4b, 0x43, 0xff}
	colorX          = color.NRGBA{0x2f, 0x6d, 0xb5, 0xff}
	colorConfetti   = []color.NRGBA{colorO, colorX, {0xe8, 0xb3, 0x3c, 0xff}, {0x4c, 0x9a, 0x6a, 0xff}}
)

func playerColor(p engine.Player) color.NRGBA {
	if p == engine.O {
		return colorO
	}
	return colorX
}

// rect is an axis aligned rectangle in virtual coordinates.
type rect struct{ x, y, w, h float64 }

func (r rect) contains(x, y float64) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

func (r rect) center() (float64, float64) { return r.x + r.w/2, r.y + r.h/2 }

// miniBoardRect returns the area of a small board, row-major from the top left.
func miniBoardRect(board int) rect {
	return rect{
		x: gridX + float64(board%3)*bigCell + miniInset,
		y: gridY + float64(board/3)*bigCell + miniInset,
		w: miniSize,
		h: miniSize,
	}
}

// bigCellCenter returns the center of the area holding a small board.
func bigCellCenter(board int) (float64, float64) {
	return gridX + (float64(board%3)+0.5)*bigCell, gridY + (float64(board/3)+0.5)*bigCell
}

// cellRect returns the area of one cell.
func cellRect(m engine.Move) rect {
	b := miniBoardRect(m.Board())
	return rect{
		x: b.x + float64(m.Cell()%3)*smallCell,
		y: b.y + float64(m.Cell()/3)*smallCell,
		w: smallCell,
		h: smallCell,
	}
}

// moveAt returns the cell under a virtual point.
func moveAt(x, y float64) (engine.Move, bool) {
	for board := range 9 {
		b := miniBoardRect(board)
		if b.contains(x, y) {
			col, row := int((x-b.x)/smallCell), int((y-b.y)/smallCell)
			return engine.MoveAt(board, min(row, 2)*3+min(col, 2)), true
		}
	}
	return engine.NoMove, false
}
