// Package ai picks moves with a Monte Carlo tree search.
//
// A Searcher runs one independent tree per thread (root parallelisation) and merges
// the statistics of the root moves: goroutines natively, Web Workers in the browser.
// Trees are kept between moves, can keep growing while the opponent thinks, and prove
// positions that are won or lost by force.
package ai

import (
	"math"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"GoTicTacToe/lib/engine"
)

const (
	// explorationConstant balances exploitation and exploration in the UCT formula.
	// Self-play at equal time showed 0.5 clearly stronger than the textbook sqrt(2).
	explorationConstant = 0.5
	// drawReward is the score of a draw for the player who moved, a win scoring 1.
	drawReward = 0.2
	// checkInterval is how many iterations run between two clock reads.
	checkInterval = 64
	// progressInterval is how often a running search reports its statistics.
	progressInterval = 100 * time.Millisecond
	// maxNodes bounds the memory of all the trees of a searcher together.
	maxNodes = 1_200_000
)

// Proof states that a move is known to win or lose whatever the opponent does.
type Proof int8

const (
	Unproven Proof = 0
	Win      Proof = 1
	Loss     Proof = -1
)

// MoveStat holds the search statistics of one root move, from the point of view of the player making it.
type MoveStat struct {
	Move   engine.Move `json:"m"`
	Visits int         `json:"v"`
	Wins   float64     `json:"w"`
	Proof  Proof       `json:"p,omitempty"`
}

// Limits bounds how long a search runs: it stops at whichever limit comes first.
type Limits struct {
	Duration    time.Duration // thinking time
	Simulations int           // playouts across all threads, 0 for no limit
	Threads     int           // threads to use, 0 or more than available for all of them
}

// Result is the outcome of a search.
type Result struct {
	Move        engine.Move
	Simulations int     // playouts in the trees, including those kept from earlier searches
	WinRate     float64 // estimated chance of winning with Move for the player to move, between 0 and 1
	Proof       Proof   // Win when Move wins by force, Loss when every move loses by force
	Threads     int
}

// Budget bounds a single tree search.
type Budget struct {
	Deadline    time.Time
	Simulations int              // playouts to run, 0 for no limit
	UntilFull   bool             // stop once the tree reaches its node limit, used for pondering
	Stop        *atomic.Bool     // optional, ends the search early when set
	Progress    func([]MoveStat) // optional, called about every progressInterval
}

type node struct {
	parent   *node
	children []*node
	untried  []engine.Move
	visits   int32
	wins     float32
	move     engine.Move
	mover    engine.Player // player who played move
	proof    Proof         // known outcome for mover
}

func newNode(parent *node, move engine.Move, state *engine.Game) *node {
	return &node{
		parent:  parent,
		move:    move,
		mover:   state.Turn().Opponent(),
		untried: state.LegalMoves(nil),
	}
}

// Worker owns one search tree and runs on a single thread. Searching a position reachable
// within two moves of its previous root reuses the matching part of the tree.
type Worker struct {
	root     *node
	state    engine.Game // position at the root
	nodes    int
	maxNodes int
	rng      *rand.Rand
	moves    []engine.Move // scratch buffer for playouts
}

// NewWorker returns a worker whose tree holds at most maxNodes nodes.
func NewWorker(maxNodes int) *Worker {
	return &Worker{maxNodes: maxNodes, rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
}

// NodesPerThread splits the node budget of a searcher between its threads.
func NodesPerThread(threads int) int { return maxNodes / max(threads, 1) }

// Search grows the tree of state within the budget. It returns the root move statistics
// and the number of playouts it ran.
func (w *Worker) Search(state engine.Game, b Budget) ([]MoveStat, int) {
	w.reroot(state)
	played := 0
	lastProgress := time.Now()
	for !state.Over() && w.root.proof == Unproven {
		if b.Simulations > 0 && played >= b.Simulations {
			break
		}
		if played%checkInterval == 0 {
			now := time.Now()
			if !now.Before(b.Deadline) || (b.Stop != nil && b.Stop.Load()) || (b.UntilFull && w.nodes >= w.maxNodes) {
				break
			}
			if b.Progress != nil && now.Sub(lastProgress) >= progressInterval {
				b.Progress(w.stats())
				lastProgress = now
			}
		}
		w.iterate()
		played++
	}
	return w.stats(), played
}

// reroot moves the root to state, keeping the matching subtree when there is one.
func (w *Worker) reroot(state engine.Game) {
	if w.root != nil {
		if n := w.find(state); n != nil {
			n.parent = nil
			w.root, w.state = n, state
			w.nodes = int(n.visits) + 1 // upper bound: each playout adds at most one node
			return
		}
	}
	w.root, w.state, w.nodes = newNode(nil, engine.NoMove, &state), state, 1
}

// find returns the node of state in the tree, looking up to two moves below the root.
func (w *Worker) find(state engine.Game) *node {
	if w.state == state {
		return w.root
	}
	for _, c := range w.root.children {
		after := w.state
		after.Play(c.move)
		if after == state {
			return c
		}
		for _, gc := range c.children {
			next := after
			next.Play(gc.move)
			if next == state {
				return gc
			}
		}
	}
	return nil
}

// iterate performs one selection, expansion, simulation and backpropagation step.
func (w *Worker) iterate() {
	game := w.state
	n := w.root

	for n.proof == Unproven && len(n.untried) == 0 && len(n.children) > 0 {
		n = n.selectChild()
		game.Play(n.move)
	}

	if n.proof == Unproven && len(n.untried) > 0 && w.nodes < w.maxNodes {
		i := w.rng.IntN(len(n.untried))
		move := n.untried[i]
		n.untried[i] = n.untried[len(n.untried)-1]
		n.untried = n.untried[:len(n.untried)-1]
		game.Play(move)
		child := newNode(n, move, &game)
		w.nodes++
		n.children = append(n.children, child)
		if game.Over() && game.Winner() == child.mover {
			child.proof = Win
			prove(n)
		}
		n = child
	}

	var winner engine.Player
	switch n.proof {
	case Win:
		winner = n.mover
	case Loss:
		winner = n.mover.Opponent()
	default:
		for !game.Over() {
			w.moves = game.LegalMoves(w.moves)
			game.Play(w.moves[w.rng.IntN(len(w.moves))])
		}
		winner = game.Winner()
	}

	for ; n != nil; n = n.parent {
		n.visits++
		switch winner {
		case n.mover:
			n.wins++
		case engine.Nobody:
			n.wins += drawReward
		}
	}
}

// prove settles n, then its ancestors, once their children decide them: a node is lost for
// its mover when the opponent has a winning reply, and won when every reply loses.
func prove(n *node) {
	for ; n != nil && n.proof == Unproven; n = n.parent {
		allLost := len(n.untried) == 0
		for _, c := range n.children {
			if c.proof == Win {
				n.proof = Loss
				break
			}
			allLost = allLost && c.proof == Loss
		}
		if n.proof == Unproven && allLost {
			n.proof = Win
		}
		if n.proof == Unproven {
			return
		}
	}
}

// selectChild returns the child with the best UCT score, never a move proven to lose.
func (n *node) selectChild() *node {
	logVisits := math.Log(float64(n.visits))
	var best *node
	bestScore := math.Inf(-1)
	for _, c := range n.children {
		if c.proof == Loss {
			continue
		}
		v := float64(c.visits)
		score := float64(c.wins)/v + explorationConstant*math.Sqrt(logVisits/v)
		if score > bestScore {
			best, bestScore = c, score
		}
	}
	return best
}

func (w *Worker) stats() []MoveStat {
	stats := make([]MoveStat, len(w.root.children))
	for i, c := range w.root.children {
		stats[i] = MoveStat{Move: c.move, Visits: int(c.visits), Wins: float64(c.wins), Proof: c.proof}
	}
	return stats
}

// Merge combines the root statistics of several trees. It picks a move proven to win if
// there is one, otherwise the most visited move not proven to lose.
func Merge(trees ...[]MoveStat) Result {
	total := map[engine.Move]*MoveStat{}
	for _, stats := range trees {
		for _, s := range stats {
			t := total[s.Move]
			if t == nil {
				t = &MoveStat{Move: s.Move}
				total[s.Move] = t
			}
			t.Visits += s.Visits
			t.Wins += s.Wins
			if s.Proof != Unproven {
				t.Proof = s.Proof // a proof holds whatever the tree that found it
			}
		}
	}

	res := Result{Move: engine.NoMove, Threads: len(trees), Proof: Loss}
	var best *MoveStat
	better := func(s *MoveStat) bool {
		switch {
		case best == nil:
			return true
		case (s.Proof == Win) != (best.Proof == Win):
			return s.Proof == Win
		case (s.Proof == Loss) != (best.Proof == Loss):
			return best.Proof == Loss
		case s.Visits != best.Visits:
			return s.Visits > best.Visits
		}
		return s.Move < best.Move
	}
	for _, s := range total {
		res.Simulations += s.Visits
		if s.Proof != Loss {
			res.Proof = Unproven
		}
		if better(s) {
			best = s
		}
	}
	if best == nil {
		return Result{Move: engine.NoMove, Threads: len(trees)}
	}
	res.Move = best.Move
	switch {
	case best.Proof == Win:
		res.Proof, res.WinRate = Win, 1
	case res.Proof == Loss:
		res.WinRate = 0
	default:
		res.WinRate = best.Wins / float64(max(best.Visits, 1))
	}
	return res
}
