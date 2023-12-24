package models

// GameSymbol determine the symbols contained in the game
type GameSymbol rune

const (
	PLAYER1 GameSymbol = 'O'
	PLAYER2 GameSymbol = 'X'
	EMPTY   GameSymbol = ' ' // for empty cell
	NONE    GameSymbol = 0
)
