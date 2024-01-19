package main

import (
	"GoTicTacToe/lib/graphics"
	"math"
	"math/rand"
	"time"
)

// Exploration constant for Monte Carlo Tree Search,
// used in UCT formula, balance between exploration and exploitation
var (
	ExplorationConstant = math.Sqrt(2)
)

// Clone a game with a deep copy of the state
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

// Node for Monte Carlo Tree Search
type Node struct {
	parent       *Node
	children     []*Node
	move         graphics.BoardCoord
	state        *Game
	visits       int
	wins         float64
	untriedMoves []graphics.BoardCoord
	playerTurn   GameSymbol
}

// Get all the possible moves for the current state of the game
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

// Runs the Monte Carlo Tree Search algorithm for a given game state and a specified time.
// returns the best move found, the number of visits and the win probability
func (g *Game) MonteCarloMove() (graphics.BoardCoord, int, float64) {
	rootMove := graphics.BoardCoord{MainBoardRow: -1, MainBoardCol: -1, MiniBoardRow: -1, MiniBoardCol: -1}
	rootNode := NewNode(nil, g, rootMove, g.playing)
	currentTime := time.Now()
	for float64(time.Since(currentTime).Milliseconds()) < g.AIDifficulty*float64(time.Second.Milliseconds()) {
		node := rootNode
		// Selection
		for !node.HasUntriedMoves() && node.HasChildren() && node.state.state == Playing {
			node = node.UCTSelectChild()
		}
		game := node.state.clone()
		// Expansion
		if node.HasUntriedMoves() && game.state == Playing {
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
			node.Update(game.GetResult(game.getOpponents(node.playerTurn)))
			node = node.parent
		}
	}

	mostVisitedChild := rootNode.MostVisitedChild()
	winProbability := mostVisitedChild.wins / float64(mostVisitedChild.visits)
	return mostVisitedChild.move, rootNode.visits, winProbability
}

// Create a new node for the Monte Carlo Tree Search and attach it to its parent
func NewNode(parent *Node, state *Game, move graphics.BoardCoord, playerTurn GameSymbol) *Node {
	node := &Node{
		parent:       parent,
		state:        state.clone(),
		move:         move,
		children:     []*Node{},
		visits:       0,
		wins:         0,
		untriedMoves: state.getPossibleMoves(),
		playerTurn:   playerTurn,
	}
	return node
}

// Check if the node has untried moves
func (n *Node) HasUntriedMoves() bool {
	return len(n.untriedMoves) > 0
}

// Get the most visited child of the node, used when returning the best move
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

// Select the best child of the node using the UCT formula
func (n *Node) UCTSelectChild() *Node {
	bestScore := math.Inf(-1)
	var bestChild *Node

	for _, child := range n.children {
		// Formula balancing exploration (of nodes with good win probabilities) and exploration (of nodes with few visits)
		uctValue := child.wins/float64(child.visits) + ExplorationConstant*math.Sqrt(math.Log(float64(n.visits))/float64(child.visits))
		if uctValue > bestScore {
			bestScore = uctValue
			bestChild = child
		}
	}

	return bestChild
}

// Get a list of all the moves not yet tried for a spceific node
func (n *Node) GetUntriedMove() graphics.BoardCoord {
	index := rand.Intn(len(n.untriedMoves))
	move := n.untriedMoves[index]
	n.untriedMoves = append(n.untriedMoves[:index], n.untriedMoves[index+1:]...)
	return move
}

// Add a child to a node
func (n *Node) AddChild(move graphics.BoardCoord, state *Game) *Node {
	child := NewNode(n, state, move, state.playing)
	n.children = append(n.children, child)
	return child
}

// Update the number of visits and wins for a node, used during backpropagation phase
func (n *Node) Update(result float64) {
	n.visits++
	n.wins += result
}

// Get the result of a game for a specific player, used during backpropagation phase
func (g *Game) GetResult(playerJustMoved GameSymbol) float64 {
	if g.win == playerJustMoved {
		return 1
	} else if g.win == NONE {
		return 0.2
	}
	return 0
}

func (n *Node) HasChildren() bool {
	return len(n.children) > 0
}
