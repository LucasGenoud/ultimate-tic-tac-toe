// Package ai picks moves with a Monte Carlo tree search.
//
// Search runs one independent tree per CPU core (root parallelisation) and merges
// the statistics of the root moves: goroutines natively, Web Workers in the browser.
package ai

import (
	"math"
	"math/rand/v2"
	"time"

	"GoTicTacToe/lib/engine"
)

const (
	// explorationConstant balances exploitation and exploration in the UCT formula.
	// Self-play at equal time showed 0.5 clearly stronger than the textbook sqrt(2).
	explorationConstant = 0.5
	// drawReward is the score of a draw for the player who moved, a win scoring 1.
	drawReward = 0.2
	// deadlineCheckInterval is how many iterations run between two clock reads.
	deadlineCheckInterval = 64
)

// MoveStat holds the search statistics of one root move.
type MoveStat struct {
	Move   engine.Move `json:"m"`
	Visits int         `json:"v"`
	Wins   float64     `json:"w"`
}

// Result is the outcome of a search.
type Result struct {
	Move        engine.Move
	Simulations int     // playouts run across all workers
	WinRate     float64 // estimated chance of winning with Move, between 0 and 1
	Workers     int     // number of parallel searches
}

type node struct {
	parent   *node
	children []*node
	untried  []engine.Move
	move     engine.Move
	mover    engine.Player // player who played move
	visits   int
	wins     float64
}

// tree is a single-threaded search tree.
type tree struct {
	root  *node
	state engine.Game
	rng   *rand.Rand
	moves []engine.Move // scratch buffer for playouts
}

func newTree(state engine.Game) *tree {
	t := &tree{state: state, rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
	t.root = t.newNode(nil, engine.NoMove, &state)
	return t
}

func (t *tree) newNode(parent *node, move engine.Move, state *engine.Game) *node {
	return &node{
		parent:  parent,
		move:    move,
		mover:   state.Turn().Opponent(),
		untried: state.LegalMoves(nil),
	}
}

// run iterates until deadline.
func (t *tree) run(deadline time.Time) {
	for i := 0; ; i++ {
		if i%deadlineCheckInterval == 0 && !time.Now().Before(deadline) {
			return
		}
		t.iterate()
	}
}

// iterate performs one selection, expansion, simulation and backpropagation step.
func (t *tree) iterate() {
	game := t.state
	n := t.root

	for len(n.untried) == 0 && len(n.children) > 0 {
		n = n.selectChild()
		game.Play(n.move)
	}

	if len(n.untried) > 0 {
		i := t.rng.IntN(len(n.untried))
		move := n.untried[i]
		n.untried[i] = n.untried[len(n.untried)-1]
		n.untried = n.untried[:len(n.untried)-1]
		game.Play(move)
		child := t.newNode(n, move, &game)
		n.children = append(n.children, child)
		n = child
	}

	for !game.Over() {
		t.moves = game.LegalMoves(t.moves)
		game.Play(t.moves[t.rng.IntN(len(t.moves))])
	}

	for ; n != nil; n = n.parent {
		n.visits++
		switch game.Winner() {
		case n.mover:
			n.wins++
		case engine.Nobody:
			n.wins += drawReward
		}
	}
}

// selectChild returns the child with the best UCT score.
func (n *node) selectChild() *node {
	logVisits := math.Log(float64(n.visits))
	best, bestScore := n.children[0], math.Inf(-1)
	for _, c := range n.children {
		v := float64(c.visits)
		score := c.wins/v + explorationConstant*math.Sqrt(logVisits/v)
		if score > bestScore {
			best, bestScore = c, score
		}
	}
	return best
}

func (t *tree) stats() []MoveStat {
	stats := make([]MoveStat, len(t.root.children))
	for i, c := range t.root.children {
		stats[i] = MoveStat{Move: c.move, Visits: c.visits, Wins: c.wins}
	}
	return stats
}

// Explore searches state on the current goroutine until deadline and returns the root move statistics.
func Explore(state engine.Game, deadline time.Time) []MoveStat {
	t := newTree(state)
	t.run(deadline)
	return t.stats()
}

// Merge combines the root statistics of several trees and picks the most visited move.
func Merge(trees ...[]MoveStat) Result {
	total := map[engine.Move]*MoveStat{}
	for _, stats := range trees {
		for _, s := range stats {
			if total[s.Move] == nil {
				total[s.Move] = &MoveStat{Move: s.Move}
			}
			total[s.Move].Visits += s.Visits
			total[s.Move].Wins += s.Wins
		}
	}

	res := Result{Move: engine.NoMove, Workers: len(trees)}
	best := -1
	for _, s := range total {
		res.Simulations += s.Visits
		if s.Visits > best || (s.Visits == best && s.Move < res.Move) {
			best, res.Move = s.Visits, s.Move
			res.WinRate = s.Wins / float64(max(s.Visits, 1))
		}
	}
	return res
}
