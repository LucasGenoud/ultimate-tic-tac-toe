package main

import (
	"GoTicTacToe/lib/graphics"
	"GoTicTacToe/lib/models"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

var (
	EXPLORATION_CONSTANT = math.Sqrt(2)
)

func (g *Game) clone() *Game {
	clonedGame := &Game{
		playing:  g.playing,
		state:    g.state,
		round:    g.round,
		pointsO:  g.pointsO,
		pointsX:  g.pointsX,
		win:      g.win,
		lastPlay: g.lastPlay,
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			clonedGame.gameBoard[i][j] = MiniBoard{
				Winner: g.gameBoard[i][j].Winner,
			}
			for k := 0; k < 3; k++ {
				for l := 0; l < 3; l++ {
					clonedGame.gameBoard[i][j].Board[k][l] = g.gameBoard[i][j].Board[k][l]
				}
			}
		}
	}

	return clonedGame
}

func (g *Game) getPossibleMoves() []graphics.BoardCoord {
	possibleMoves := make([]graphics.BoardCoord, 0)

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.isValidPlay(i, j) && g.gameBoard[i][j].Winner == EMPTY {
				for k := 0; k < 3; k++ {
					for l := 0; l < 3; l++ {
						coord := graphics.BoardCoord{MainBoardRow: i, MainBoardCol: j, MiniBoardRow: k, MiniBoardCol: l}
						if g.getValueOfCoordinates(coord) == EMPTY {
							possibleMoves = append(possibleMoves, coord)
						}
					}
				}
			}

		}
	}
	return possibleMoves
}
func (g *Game) MonteCarloMove() (graphics.BoardCoord, int, float64) {
	var wg sync.WaitGroup
	numCPU := runtime.NumCPU()
	results := make(chan *Node, numCPU)

	for i := 0; i < 1; i++ {
		wg.Add(1)
		go g.MonteCarloTreeSearch(&wg, results)
	}

	wg.Wait()
	close(results)

	var bestMove *Node
	var totalVisits int
	for result := range results {
		totalVisits += result.visits
		if bestMove == nil || result.MostVisitedChild().visits > bestMove.MostVisitedChild().visits {
			bestMove = result
		}
	}

	winProbability := bestMove.MostVisitedChild().wins / float64(bestMove.MostVisitedChild().visits)
	return bestMove.MostVisitedChild().move, totalVisits, winProbability
}

func (g *Game) MonteCarloTreeSearch(wg *sync.WaitGroup, results chan *Node) {
	defer wg.Done()
	rootMove := graphics.BoardCoord{MainBoardRow: -1, MainBoardCol: -1, MiniBoardRow: -1, MiniBoardCol: -1}
	rootNode := NewNode(nil, g, rootMove, g.playing)
	currentTime := time.Now()
	for time.Since(currentTime).Milliseconds() < 5000 {
		node := rootNode
		game := g.clone()
		// Selection and Expansion
		for node.HasUntriedMoves() == false && node.HasChildren() {
			node = node.UCTSelectChild()
			game.makePlay(node.move)
		}
		// Expand the node (if possible)
		if node.HasUntriedMoves() {
			move := node.GetUntriedMove()
			game.makePlay(move)
			node = node.AddChild(move, game)
		}
		// Simulation
		for game.state == Playing {
			possibleMoves := game.getPossibleMoves()
			randomMove := possibleMoves[rand.Intn(len(possibleMoves))]
			game.makePlay(randomMove)
		}
		// Backpropagation
		for node != nil {
			node.Update(game.GetResult(game.getOpponents(node.playerJustMoved)))
			node = node.parent
		}
	}
	results <- rootNode
}

type Node struct {
	parent          *Node
	children        []*Node
	move            graphics.BoardCoord
	state           *Game
	visits          int
	wins            float64
	untriedMoves    []graphics.BoardCoord
	playerJustMoved models.GameSymbol
}

func NewNode(parent *Node, state *Game, move graphics.BoardCoord, playerJustMoved models.GameSymbol) *Node {
	node := &Node{
		parent:          parent,
		state:           state.clone(),
		move:            move,
		children:        []*Node{},
		visits:          0,
		wins:            0,
		untriedMoves:    state.getPossibleMoves(),
		playerJustMoved: playerJustMoved,
	}
	return node
}

func (n *Node) HasUntriedMoves() bool {
	return len(n.untriedMoves) > 0
}
func (n *Node) MostVisitedChild() *Node {
	mostVisits := -1
	var mostVisitedChild *Node

	for _, child := range n.children {
		if child.visits > mostVisits {
			mostVisits = child.visits
			mostVisitedChild = child
		}
	}

	return mostVisitedChild
}
func (n *Node) UCTSelectChild() *Node {
	bestScore := math.Inf(-1)
	var bestChild *Node

	for _, child := range n.children {
		uctValue := child.wins/float64(child.visits) + EXPLORATION_CONSTANT*math.Sqrt(math.Log(float64(n.visits))/float64(child.visits))
		if uctValue > bestScore {
			bestScore = uctValue
			bestChild = child
		}
	}

	return bestChild
}

func (n *Node) GetUntriedMove() graphics.BoardCoord {
	index := rand.Intn(len(n.untriedMoves))
	move := n.untriedMoves[index]
	n.untriedMoves = append(n.untriedMoves[:index], n.untriedMoves[index+1:]...)
	return move
}

func (n *Node) AddChild(move graphics.BoardCoord, state *Game) *Node {
	child := NewNode(n, state, move, state.playing)
	n.children = append(n.children, child)
	return child
}

func (n *Node) Update(result float64) {
	n.visits++
	n.wins += result
}

func (g *Game) GetResult(playerJustMoved models.GameSymbol) float64 {
	if g.win == playerJustMoved {
		return 1
	} else if g.win == NONE {
		return 0
	}
	return 0
}

func (n *Node) HasChildren() bool {
	return len(n.children) > 0
}