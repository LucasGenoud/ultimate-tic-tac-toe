package main

import (
	"GoTicTacToe/lib/graphics"
	"testing"
)

func TestGameInit(t *testing.T) {
	game := &Game{}

	game.init()

	if game.playing != PLAYER1 && game.playing != PLAYER2 {
		t.Errorf("Unexpected player: %v", game.playing)
	}

	if game.state != WaitingForGameStart {
		t.Errorf("Unexpected state: %v", game.state)
	}

	if game.round != 0 {
		t.Errorf("Unexpected round: %d", game.round)
	}

	if game.win != EMPTY {
		t.Errorf("Unexpected win: %v", game.win)
	}

	if game.AIEnabled != true {
		t.Errorf("Unexpected AIEnabled: %v", game.AIEnabled)
	}
}

func TestIsValidPlay(t *testing.T) {
	game := &Game{}
	game.init()

	// Test when lastPlay is -1
	if !game.isValidPlay(0, 0) {
		t.Errorf("Expected true, got false")
	}

	// Test when lastPlay is not -1
	game.lastPlay = graphics.BoardCoord{MainBoardRow: 0, MainBoardCol: 0, MiniBoardRow: 0, MiniBoardCol: 0}
	if !game.isValidPlay(0, 0) {
		t.Errorf("Expected true, got false")
	}

	if game.isValidPlay(1, 1) {
		t.Errorf("Expected false, got true")
	}
}

func TestGetValueOfCoordinates(t *testing.T) {
	game := &Game{}
	game.init()

	coordinates := graphics.BoardCoord{MainBoardRow: 0, MainBoardCol: 0, MiniBoardRow: 0, MiniBoardCol: 0}
	game.setValueOfCoordinates(coordinates, PLAYER1)

	if game.getValueOfCoordinates(coordinates) != PLAYER1 {
		t.Errorf("Expected %v, got %v", PLAYER1, game.getValueOfCoordinates(coordinates))
	}
}

func TestSetValueOfCoordinates(t *testing.T) {
	game := &Game{}
	game.init()

	coordinates := graphics.BoardCoord{MainBoardRow: 0, MainBoardCol: 0, MiniBoardRow: 0, MiniBoardCol: 0}
	game.setValueOfCoordinates(coordinates, PLAYER1)

	if game.gameBoard[0][0].Board[0][0] != PLAYER1 {
		t.Errorf("Expected %v, got %v", PLAYER1, game.gameBoard[0][0].Board[0][0])
	}
}

func TestGetMiniBoardCoordinates(t *testing.T) {
	game := &Game{}
	game.init()

	coordinates := game.getMiniBoardCoordinates(100, 100)

	if coordinates.MainBoardRow != 0 || coordinates.MainBoardCol != 0 || coordinates.MiniBoardRow != 0 || coordinates.MiniBoardCol != 0 {
		t.Errorf("Expected (0,0,0,0), got (%v,%v,%v,%v)", coordinates.MainBoardRow, coordinates.MainBoardCol, coordinates.MiniBoardRow, coordinates.MiniBoardCol)
	}
}

func TestCheckWin(t *testing.T) {
	game := &Game{}
	game.init()

	// Set up a winning condition for PLAYER1
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			game.gameBoard[i][j].Winner = PLAYER1
		}
	}

	if game.CheckWin() != PLAYER1 {
		t.Errorf("Expected %v, got %v", PLAYER1, game.CheckWin())
	}
}
