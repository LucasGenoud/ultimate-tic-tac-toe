package main

import (
	"testing"
)

func initGame() *Game {
	game := &Game{}
	game.init()
	game.state = Playing
	game.AIDifficulty = 1

	return game
}
func TestMonteCarloMove(t *testing.T) {

	game := initGame()
	move, visits, winProbability := game.MonteCarloMove()

	if visits < 0 {
		t.Errorf("Unexpected number of visits: %d", visits)
	}

	if winProbability < 0 || winProbability > 1 {
		t.Errorf("Unexpected win probability: %f", winProbability)
	}

	if !game.isValidPlay(move.MainBoardRow, move.MainBoardCol) {
		t.Errorf("Invalid move generated: %v", move)
	}
}

func TestSimulateGame(t *testing.T) {
	game := initGame()
	for game.state == Playing {
		possibleMoves := game.getPossibleMoves()
		randomMove := possibleMoves[0]
		game.makePlay(randomMove)
	}
}

func TestAIWinAgainstRandom(t *testing.T) {
	for i := 0; i < 100; i++ {
		game := initGame()
		game.AIDifficulty = 0.1
		for game.state == Playing {
			if game.playing == PLAYER1 {
				move, _, _ := game.MonteCarloMove()
				game.makePlay(move)
			} else {
				possibleMoves := game.getPossibleMoves()
				randomMove := possibleMoves[0]
				game.makePlay(randomMove)
			}
		}
		if game.win != PLAYER1 {
			t.Errorf("AI lost against random")
		}
	}

}
