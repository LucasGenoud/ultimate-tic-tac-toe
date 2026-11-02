package ai

import (
	"testing"
	"time"

	"GoTicTacToe/lib/engine"
)

func decode(t *testing.T, text string) engine.Game {
	t.Helper()
	var g engine.Game
	if err := g.UnmarshalText([]byte(text)); err != nil {
		t.Fatal(err)
	}
	return g
}

// winInOne is a position where O, to play in board 2, wins the game with board 2 cell 2.
const winInOne = "OOO......" + "OOO......" + "OO......." +
	"..X.X.X.." + "X...X...." + "X........" +
	"........." + "........." + "........." +
	"O" + "29"

func TestSearchReturnsLegalMove(t *testing.T) {
	g := engine.New(engine.X)
	g.Play(engine.MoveAt(4, 4))

	res := NewSearcher().Search(g, Limits{Duration: 200 * time.Millisecond})

	if !g.Legal(res.Move) {
		t.Fatalf("illegal move %d", res.Move)
	}
	if res.Simulations <= 0 {
		t.Errorf("simulations = %d, want > 0", res.Simulations)
	}
	if res.WinRate < 0 || res.WinRate > 1 {
		t.Errorf("win rate = %f, want within [0, 1]", res.WinRate)
	}
	if res.Threads < 1 {
		t.Errorf("threads = %d, want >= 1", res.Threads)
	}
}

func TestSearchProvesImmediateWin(t *testing.T) {
	g := decode(t, winInOne)

	start := time.Now()
	res := NewSearcher().Search(g, Limits{Duration: 5 * time.Second})

	if want := engine.MoveAt(2, 2); res.Move != want {
		t.Errorf("search played board %d cell %d, want board 2 cell 2", res.Move.Board(), res.Move.Cell())
	}
	if res.Proof != Win || res.WinRate != 1 {
		t.Errorf("proof = %v, win rate = %v, want a proven win", res.Proof, res.WinRate)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("a proven position should end the search early, took %v", elapsed)
	}
}

func TestSimulationLimit(t *testing.T) {
	s := NewSearcher()
	res := s.Search(engine.New(engine.O), Limits{Duration: 10 * time.Second, Simulations: 1000, Threads: 2})

	if res.Threads != 2 {
		t.Errorf("threads = %d, want 2", res.Threads)
	}
	if res.Simulations < 1000 || res.Simulations > 1000+2*checkInterval {
		t.Errorf("simulations = %d, want about 1000", res.Simulations)
	}
}

func TestWorkerReusesTree(t *testing.T) {
	w := NewWorker(NodesPerThread(1))
	g := engine.New(engine.O)
	w.Search(g, Budget{Deadline: time.Now().Add(time.Hour), Simulations: 20000})

	// Two moves later, the subtree of that line is kept.
	g.Play(w.root.selectChild().move)
	g.Play(w.root.children[0].children[0].move)
	before := w.find(g)
	if before == nil {
		t.Skip("the line was not expanded two levels deep")
	}
	kept := int(before.visits)
	stats, played := w.Search(g, Budget{Deadline: time.Now().Add(time.Hour), Simulations: 100})

	total := 0
	for _, s := range stats {
		total += s.Visits
	}
	if played != 100 || total < kept+100-1 {
		t.Errorf("root moves have %d visits after %d playouts, want at least %d kept + 100", total, played, kept)
	}
}

func TestCancelStopsSearch(t *testing.T) {
	s := NewSearcher()
	done := make(chan Result)
	go func() { done <- s.Search(engine.New(engine.O), Limits{Duration: time.Minute}) }()
	time.Sleep(100 * time.Millisecond)
	s.Cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("search did not stop after Cancel")
	}
}

func TestPonderThenSearch(t *testing.T) {
	s := NewSearcher()
	g := engine.New(engine.O)
	s.Ponder(g, 0)
	time.Sleep(100 * time.Millisecond)

	res := s.Search(g, Limits{Duration: 10 * time.Millisecond})

	if !g.Legal(res.Move) {
		t.Fatalf("illegal move %d", res.Move)
	}
	if res.Simulations < 1000 {
		t.Errorf("simulations = %d, want the playouts from pondering to be kept", res.Simulations)
	}
}

func TestSearchBeatsRandomPlayer(t *testing.T) {
	if testing.Short() {
		t.Skip("plays full games")
	}
	s := NewSearcher()
	var buf []engine.Move
	for i := range 6 {
		g := engine.New(engine.O)
		for !g.Over() {
			if g.Turn() == engine.O {
				g.Play(s.Search(g, Limits{Duration: 50 * time.Millisecond}).Move)
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

func TestMerge(t *testing.T) {
	res := Merge(
		[]MoveStat{{Move: 1, Visits: 10, Wins: 5}, {Move: 2, Visits: 30, Wins: 3}},
		[]MoveStat{{Move: 1, Visits: 40, Wins: 30}, {Move: 2, Visits: 5, Wins: 1}},
	)
	if res.Move != 1 || res.Simulations != 85 || res.Threads != 2 || res.WinRate != 0.7 {
		t.Errorf("Merge = %+v, want move 1, 85 simulations, 2 threads, 0.7 win rate", res)
	}

	res = Merge(
		[]MoveStat{{Move: 1, Visits: 90, Wins: 50}, {Move: 2, Visits: 3, Wins: 3, Proof: Win}},
		[]MoveStat{{Move: 1, Visits: 90, Wins: 50, Proof: Loss}, {Move: 3, Visits: 9, Wins: 1}},
	)
	if res.Move != 2 || res.Proof != Win || res.WinRate != 1 {
		t.Errorf("Merge = %+v, want the proven win, move 2", res)
	}

	res = Merge([]MoveStat{{Move: 1, Visits: 9, Wins: 0, Proof: Loss}, {Move: 2, Visits: 4, Proof: Loss}})
	if res.Move != 1 || res.Proof != Loss || res.WinRate != 0 {
		t.Errorf("Merge = %+v, want a proven loss on the most visited move", res)
	}
}
