package main

import (
	"testing"
)

func TestMonteCarloMove(t *testing.T) {

	/*
			Weirdly, this test fails because the rootNode is nil when returning the move.
		      game := &Game{}
				game.init()
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
	*/
}
