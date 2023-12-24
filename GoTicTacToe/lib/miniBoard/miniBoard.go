package miniBoard

import (
	"GoTicTacToe/lib/models"
)

type MiniBoard struct {
	Board  [3][3]models.GameSymbol
	Winner models.GameSymbol
}

const (
	PLAYER1 models.GameSymbol = 'O'
	PLAYER2 models.GameSymbol = 'X'
	EMPTY   models.GameSymbol = ' ' // for empty cell
	NONE    models.GameSymbol = 0
)

func (g *MiniBoard) CheckWin() {
	for i := 0; i < 3; i++ {
		if g.winnerOnLine(i, 0, 0, 1) != EMPTY {
			g.Winner = g.winnerOnLine(i, 0, 0, 1)
		}

		if g.winnerOnLine(0, i, 1, 0) != EMPTY {
			g.Winner = g.winnerOnLine(0, i, 1, 0)
		}
	}
	if g.winnerOnLine(0, 0, 1, 1) != EMPTY {
		g.Winner = g.winnerOnLine(0, 0, 1, 1)
	}
	if g.winnerOnLine(0, 2, 1, -1) != EMPTY {
		g.Winner = g.winnerOnLine(0, 2, 1, -1)
	}

}

func (g *MiniBoard) winnerOnLine(x, y, dx, dy int) models.GameSymbol {
	for i := 0; i < 3; i++ {
		if g.Board[x][y] != g.Board[x+dx*i][y+dy*i] {
			return EMPTY
		}
	}
	return g.Board[x][y]
}
