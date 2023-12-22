package graphics

import (
	"github.com/fogleman/gg"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	numberOfRows       = 3
	symbolLineWidth    = 10
	mainBoardLineWidth = 10
	miniBoardLineWidth = 5
	miniBoardPadding   = 10
)

var (
	boardSize     = 300
	miniBoardSize = boardSize / numberOfRows
	symbolSize    = miniBoardSize / numberOfRows
)

type GameGraphics struct {
	MainBoard *ebiten.Image
	MiniBoard *ebiten.Image
	Circle    *ebiten.Image
	Cross     *ebiten.Image
}

func Init(boardWidth int) GameGraphics {
	boardSize = boardWidth
	miniBoardSize = boardSize/numberOfRows - miniBoardPadding*2 - miniBoardLineWidth - mainBoardLineWidth
	symbolSize = miniBoardSize / numberOfRows
	gameGraphics := GameGraphics{}
	gameGraphics.Circle = drawCircle()
	gameGraphics.Cross = drawCross()
	gameGraphics.MainBoard = DrawMainBoard()
	gameGraphics.MiniBoard = DrawMiniBoard()
	return gameGraphics
}

func DrawMainBoard() *ebiten.Image {
	ggm := gameGraphicMaker{gg.NewContext(boardSize, boardSize)}
	ggm.setRGBA(1, 1, 1, 1)
	boardCaseSize := boardSize / numberOfRows
	for i := 1; i < numberOfRows; i++ {
		ggm.drawRectangle(boardCaseSize*i-mainBoardLineWidth/2, 0, mainBoardLineWidth, boardSize)
		ggm.drawRectangle(0, boardCaseSize*i-mainBoardLineWidth/2, boardSize, mainBoardLineWidth)
	}
	ggm.fill()
	return ggm.getImage()
}

func DrawMiniBoard() *ebiten.Image {
	ggm := gameGraphicMaker{gg.NewContext(boardSize, boardSize)}
	ggm.setRGBA(1, 1, 1, 0.3)
	boardCaseSize := boardSize / (numberOfRows * numberOfRows)
	for i := 1; i < numberOfRows; i++ {
		ggm.drawRectangle(boardCaseSize*i-miniBoardLineWidth/2+miniBoardPadding, miniBoardPadding, miniBoardLineWidth, miniBoardSize-miniBoardPadding*2)
		ggm.drawRectangle(miniBoardPadding, boardCaseSize*i-miniBoardLineWidth/2+miniBoardPadding, miniBoardSize-miniBoardPadding*2, miniBoardLineWidth)
	}
	ggm.fill()
	return ggm.getImage()
}

func drawCircle() *ebiten.Image {

	var radius = symbolSize/2 - miniBoardLineWidth*2
	ggm := gameGraphicMaker{gg.NewContext(symbolSize, symbolSize)}
	ggm.setRGBA(1, 1, 1, 1)
	ggm.setLineWidth(symbolLineWidth)
	ggm.drawCircle(symbolSize/2, symbolSize/2, radius)
	ggm.stroke()

	return ggm.getImage()
}

func drawCross() *ebiten.Image {
	ggm := gameGraphicMaker{gg.NewContext(symbolSize, symbolSize)}
	ggm.setRGBA(1, 1, 1, 1)
	ggm.rotateAbout(45, symbolSize/2, symbolSize/2)
	ggm.rotateAbout(45, symbolSize/2, symbolSize/2)
	ggm.drawRectangle(symbolSize/2-symbolLineWidth/2, 0, symbolLineWidth, symbolSize)
	ggm.drawRectangle(0, symbolSize/2-symbolLineWidth/2, symbolSize, symbolLineWidth)
	ggm.fill()
	return ggm.getImage()
}

func GetPositionOfSymbol(x, y int) (float64, float64) {
	// TODO: probably need to add some padding and account for the offset depending on in which mini board the symbol is
	return float64(symbolSize*x-miniBoardLineWidth*x/2) + miniBoardPadding, float64(symbolSize*y) + miniBoardPadding
}
