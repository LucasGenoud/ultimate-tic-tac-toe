package engine

import (
	"math/rand/v2"
	"testing"
)

func TestFirstMoveCanBePlayedAnywhere(t *testing.T) {
	g := New(O)
	if got := len(g.LegalMoves(nil)); got != 81 {
		t.Fatalf("legal moves at start = %d, want 81", got)
	}
}

func TestMoveSendsOpponentToMatchingBoard(t *testing.T) {
	g := New(O)
	if !g.Play(MoveAt(0, 4)) {
		t.Fatal("first move rejected")
	}
	if g.Turn() != X {
		t.Errorf("turn = %v, want X", g.Turn())
	}
	if g.Cell(MoveAt(0, 4)) != O {
		t.Errorf("cell = %v, want O", g.Cell(MoveAt(0, 4)))
	}
	if !g.BoardPlayable(4) || g.BoardPlayable(0) {
		t.Error("only board 4 should be playable")
	}
	if g.Play(MoveAt(0, 0)) {
		t.Error("move outside the forced board was accepted")
	}
	if g.Play(MoveAt(0, 4)) {
		t.Error("move on an occupied cell was accepted")
	}
}

// playAll plays moves alternately, failing the test on an illegal one.
func playAll(t *testing.T, g *Game, moves ...Move) {
	t.Helper()
	for _, m := range moves {
		if !g.Play(m) {
			t.Fatalf("move %d (board %d, cell %d) rejected", m, m.Board(), m.Cell())
		}
	}
}

func TestWinningSmallBoardFreesNextMove(t *testing.T) {
	g := New(O)
	// O takes the top row of board 0 while X answers inside boards 1 and 2.
	playAll(t, &g,
		MoveAt(0, 1), MoveAt(1, 0),
		MoveAt(0, 2), MoveAt(2, 0),
		MoveAt(0, 0),
	)
	if g.BoardWinner(0) != O {
		t.Fatalf("board 0 winner = %v, want O", g.BoardWinner(0))
	}
	if line, ok := g.BoardLine(0); !ok || line != [3]int{0, 1, 2} {
		t.Errorf("board line = %v %v, want [0 1 2]", line, ok)
	}
	// X is sent to board 0, which is decided, so X may play in any other open board.
	if g.BoardPlayable(0) {
		t.Error("a won board must not be playable")
	}
	for b := 1; b < 9; b++ {
		if !g.BoardPlayable(b) {
			t.Errorf("board %d should be playable", b)
		}
	}
}

func TestRandomGamesTerminate(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	var buf []Move
	for range 200 {
		g := New(O)
		for !g.Over() {
			buf = g.LegalMoves(buf)
			if len(buf) == 0 {
				t.Fatal("no legal moves in an unfinished game")
			}
			playAll(t, &g, buf[rng.IntN(len(buf))])
		}
		if w := g.Winner(); w != Nobody {
			if _, ok := g.WinningLine(); !ok {
				t.Fatalf("winner %v without a winning line", w)
			}
		}
	}
}

func TestTextRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	var buf []Move
	for range 50 {
		g := New(X)
		for n := rng.IntN(60); n > 0 && !g.Over(); n-- {
			buf = g.LegalMoves(buf)
			g.Play(buf[rng.IntN(len(buf))])
		}
		text, err := g.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		var decoded Game
		if err := decoded.UnmarshalText(text); err != nil {
			t.Fatalf("decode %q: %v", text, err)
		}
		if decoded != g {
			t.Fatalf("round trip mismatch for %q", text)
		}
	}
}
