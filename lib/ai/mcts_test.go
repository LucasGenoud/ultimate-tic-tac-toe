package ai

import (
	"testing"
	"time"

	"GoTicTacToe/lib/engine"
)

func TestSearchReturnsLegalMove(t *testing.T) {
	g := engine.New(engine.X)
	g.Play(engine.MoveAt(4, 4))

	res := Search(g, 200*time.Millisecond)

	if !g.Legal(res.Move) {
		t.Fatalf("illegal move %d", res.Move)
	}
	if res.Simulations <= 0 {
		t.Errorf("simulations = %d, want > 0", res.Simulations)
	}
	if res.WinRate < 0 || res.WinRate > 1 {
		t.Errorf("win rate = %f, want within [0, 1]", res.WinRate)
	}
	if res.Workers < 1 {
		t.Errorf("workers = %d, want >= 1", res.Workers)
	}
}

func TestSearchTakesImmediateWin(t *testing.T) {
	// O owns boards 0 and 1 and two cells of board 2, where X's last move (board 3, cell 2) sends O.
	position := "OOO......" + "OOO......" + "OO......." +
		"..X.X.X.." + "X...X...." + "X........" +
		"........." + "........." + "........." +
		"O" + "29"
	var g engine.Game
	if err := g.UnmarshalText([]byte(position)); err != nil {
		t.Fatal(err)
	}

	res := Search(g, 200*time.Millisecond)

	if want := engine.MoveAt(2, 2); res.Move != want {
		t.Errorf("search played board %d cell %d, want the winning move board 2 cell 2", res.Move.Board(), res.Move.Cell())
	}
}

func TestSearchBeatsRandomPlayer(t *testing.T) {
	if testing.Short() {
		t.Skip("plays full games")
	}
	var buf []engine.Move
	for i := range 6 {
		g := engine.New(engine.O)
		for !g.Over() {
			if g.Turn() == engine.O {
				g.Play(Search(g, 50*time.Millisecond).Move)
			} else {
				buf = g.LegalMoves(buf)
				g.Play(buf[i%len(buf)])
			}
		}
		if g.Winner() != engine.O {
			t.Errorf("game %d: winner = %v, want O", i, g.Winner())
		}
	}
}

func TestMergeSumsTrees(t *testing.T) {
	res := Merge(
		[]MoveStat{{Move: 1, Visits: 10, Wins: 5}, {Move: 2, Visits: 30, Wins: 3}},
		[]MoveStat{{Move: 1, Visits: 40, Wins: 30}, {Move: 2, Visits: 5, Wins: 1}},
	)
	if res.Move != 1 || res.Simulations != 85 || res.Workers != 2 || res.WinRate != 0.7 {
		t.Errorf("Merge = %+v, want move 1, 85 simulations, 2 workers, 0.7 win rate", res)
	}
}
