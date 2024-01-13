package main

import "GoTicTacToe/lib/graphics"

// GameSymbol determine the symbols contained in the game
type GameSymbol rune
type GameState int

type Game struct {
	playing          GameSymbol                                // current player as a symbol
	state            GameState                                 // current state of the game
	gameBoard        [BoardRowLength][BoardRowLength]MiniBoard // the game board
	round            int                                       // current round index
	pointsO          int                                       // points of player 1
	pointsX          int                                       // points of player 2
	win              GameSymbol                                // winner symbol or EMPTY if the game has not ended yet
	lastPlay         graphics.BoardCoord                       // last play coordinates
	AISimulations    int                                       // number of simulations done by the AI
	AIWinProbability float64                                   // probability of winning for the AI
	AIRunning        bool                                      // true if the AI is processing a move
	AIDifficulty     float64                                   // difficulty level of the AI
	AIEnabled        bool                                      // true if the AI is enabled
}
