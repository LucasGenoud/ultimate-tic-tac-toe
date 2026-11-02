package main

import (
	"image/color"

	"GoTicTacToe/lib/engine"
)

// Everything is laid out in a virtual 800 x 1010 space, scaled to fit the window.
const (
	viewWidth  = 800.0
	viewHeight = 1010.0

	boardX       = 40.0
	boardY       = 128.0
	boardSize    = 720.0
	boardPadding = 20.0
	gridX        = boardX + boardPadding
	gridY        = boardY + boardPadding
	gridSize     = boardSize - 2*boardPadding
	bigCell      = gridSize / 3
	miniInset    = 14.0
	miniSize     = bigCell - 2*miniInset
	smallCell    = miniSize / 3

	titleY      = 50.0
	statusY     = 92.0
	statsLabelY = 876.0
	statsValueY = 902.0
	barY        = 924.0
	buttonsY    = 944.0
	buttonH     = 48.0
)

// palette holds the colors of a theme.
type palette struct {
	background, surface, border color.NRGBA
	ink, muted                  color.NRGBA
	gridMajor, gridMinor        color.NRGBA
	shadow                      color.NRGBA
	o, x                        color.NRGBA
	confetti                    []color.NRGBA
}

var (
	lightPalette = palette{
		background: color.NRGBA{0xf4, 0xf2, 0xee, 0xff},
		surface:    color.NRGBA{0xff, 0xff, 0xff, 0xff},
		border:     color.NRGBA{0xe2, 0xdf, 0xd8, 0xff},
		ink:        color.NRGBA{0x24, 0x26, 0x2b, 0xff},
		muted:      color.NRGBA{0x6e, 0x71, 0x7a, 0xff},
		gridMajor:  color.NRGBA{0x3a, 0x3d, 0x44, 0xff},
		gridMinor:  color.NRGBA{0xdc, 0xd8, 0xd0, 0xff},
		shadow:     color.NRGBA{0x00, 0x00, 0x00, 0x14},
		o:          color.NRGBA{0xcf, 0x4b, 0x43, 0xff},
		x:          color.NRGBA{0x2f, 0x6d, 0xb5, 0xff},
		confetti:   []color.NRGBA{{0xcf, 0x4b, 0x43, 0xff}, {0x2f, 0x6d, 0xb5, 0xff}, {0xe8, 0xb3, 0x3c, 0xff}, {0x4c, 0x9a, 0x6a, 0xff}},
	}
	darkPalette = palette{
		background: color.NRGBA{0x17, 0x18, 0x1b, 0xff},
		surface:    color.NRGBA{0x22, 0x24, 0x28, 0xff},
		border:     color.NRGBA{0x34, 0x36, 0x3c, 0xff},
		ink:        color.NRGBA{0xe8, 0xe6, 0xe1, 0xff},
		muted:      color.NRGBA{0x9a, 0x9c, 0xa3, 0xff},
		gridMajor:  color.NRGBA{0xb4, 0xb2, 0xac, 0xff},
		gridMinor:  color.NRGBA{0x3a, 0x3d, 0x43, 0xff},
		shadow:     color.NRGBA{0x00, 0x00, 0x00, 0x50},
		o:          color.NRGBA{0xe4, 0x6a, 0x62, 0xff},
		x:          color.NRGBA{0x5b, 0x95, 0xd6, 0xff},
		confetti:   []color.NRGBA{{0xe4, 0x6a, 0x62, 0xff}, {0x5b, 0x95, 0xd6, 0xff}, {0xe8, 0xb3, 0x3c, 0xff}, {0x5f, 0xb0, 0x7d, 0xff}},
	}
)

func (p *palette) player(pl engine.Player) color.NRGBA {
	if pl == engine.O {
		return p.o
	}
	return p.x
}

// rect is an axis aligned rectangle in virtual coordinates.
type rect struct{ x, y, w, h float64 }

func (r rect) contains(x, y float64) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

func (r rect) center() (float64, float64) { return r.x + r.w/2, r.y + r.h/2 }

// row splits r horizontally into n rectangles separated by gap.
func (r rect) row(n int, gap float64) []rect {
	w := (r.w - gap*float64(n-1)) / float64(n)
	cells := make([]rect, n)
	for i := range cells {
		cells[i] = rect{x: r.x + float64(i)*(w+gap), y: r.y, w: w, h: r.h}
	}
	return cells
}

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

// Footer buttons.
var (
	hintButton    = rect{x: boardX, y: buttonsY, w: 110, h: buttonH}
	undoButton    = rect{x: boardX + 122, y: buttonsY, w: 110, h: buttonH}
	newGameButton = rect{x: boardX + boardSize - 132, y: buttonsY, w: 132, h: buttonH}
)
