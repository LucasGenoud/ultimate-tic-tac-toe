package main

import (
	"GoTicTacToe/lib/graphics"
	"math/rand"
)

func (g *Game) MonteCarloMove() graphics.BoardCoord {
	// Number of simulations
	numSimulations := 200

	// Store the win rates for each possible move
	winRates := make(map[graphics.BoardCoord]float64)

	possibleMoves := g.getPossibleMoves()

	// Simulate each possible move
	for _, move := range possibleMoves {
		winRates[move] = g.simulateGames(move, numSimulations)
	}

	// Find the move with the highest win rate
	bestMove := graphics.BoardCoord{}
	bestRate := -1.0
	for move, rate := range winRates {
		if rate > bestRate {
			bestMove = move
			bestRate = rate
		}
	}

	return bestMove
}
func (g *Game) simulateGames(startMove graphics.BoardCoord, numSimulations int) float64 {
	wins := 0

	for i := 0; i < numSimulations; i++ {
		// Clone the current game state
		clonedGame := g.clone()

		// Make the initial move
		clonedGame.setValueOfCoordinates(startMove, clonedGame.playing)
		clonedGame.switchPlayer()

		// Play out the rest of the game randomly
		for clonedGame.state == Playing {
			possibleMoves := clonedGame.getPossibleMoves()
			randomMove := possibleMoves[rand.Intn(len(possibleMoves))]
			clonedGame.setValueOfCoordinates(randomMove, clonedGame.playing)
			clonedGame.switchPlayer()
			clonedGame.wins(clonedGame.CheckWin())
		}

		// Check the outcome
		if clonedGame.win == g.playing {
			wins++
		}
	}

	return float64(wins) / float64(numSimulations)
}

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
	var possibleMoves []graphics.BoardCoord = make([]graphics.BoardCoord, 0)

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.gameBoard[i][j].Winner == EMPTY && g.isValidPlay(i, j) {
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
