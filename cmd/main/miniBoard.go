package main

type MiniBoard struct {
	Board  [3][3]GameSymbol
	Winner GameSymbol
}

func (g *MiniBoard) CheckWin() {
	for i := 0; i < 3; i++ {
		if g.winnerOnLine(i, 0, 0, 1) != EMPTY {
			g.Winner = g.winnerOnLine(i, 0, 0, 1)
			return
		}

		if g.winnerOnLine(0, i, 1, 0) != EMPTY {
			g.Winner = g.winnerOnLine(0, i, 1, 0)
			return
		}
	}
	if g.winnerOnLine(0, 0, 1, 1) != EMPTY {
		g.Winner = g.winnerOnLine(0, 0, 1, 1)
		return
	}
	if g.winnerOnLine(0, 2, 1, -1) != EMPTY {
		g.Winner = g.winnerOnLine(0, 2, 1, -1)
		return
	}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.Board[i][j] == EMPTY {
				return
			}
		}
	}
	g.Winner = NONE
}

func (g *MiniBoard) winnerOnLine(x, y, dx, dy int) GameSymbol {
	for i := 0; i < 3; i++ {
		if g.Board[x][y] != g.Board[x+dx*i][y+dy*i] {
			return EMPTY
		}
	}
	return g.Board[x][y]
}
