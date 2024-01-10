package main

import (
	"fmt"
	"testing"
)

func TestMonteCarloMove(t *testing.T) {
	game := &Game{}
	game.init()
	move, visits, winProbability := game.MonteCarloMove()

	fmt.Printf("Move: %v, Visits: %d, Win Probability: %f\n", move, visits, winProbability)

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
