// Package engine implements the rules of ultimate tic-tac-toe, independently of any rendering.
package engine

import "fmt"

// Player identifies who owns a cell or won a board.
type Player uint8

const (
	Nobody Player = iota
	O
	X
)

// Opponent returns the other player.
func (p Player) Opponent() Player {
	if p == O {
		return X
	}
	return O
}

func (p Player) String() string {
	switch p {
	case O:
		return "O"
	case X:
		return "X"
	}
	return " "
}

// Move is a cell index on the whole grid: board*9 + cell, where board and cell are row*3 + col.
type Move int8

// NoMove is returned when there is no move to report.
const NoMove Move = -1

// MoveAt builds a move from a small board index and a cell index within that board.
func MoveAt(board, cell int) Move { return Move(board*9 + cell) }

// Board returns the small board the move is played in.
func (m Move) Board() int { return int(m) / 9 }

// Cell returns the cell of the small board the move is played in.
func (m Move) Cell() int { return int(m) % 9 }

// Lines lists the eight ways to align three cells on a 3x3 grid.
var Lines = [8][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
	{0, 4, 8}, {2, 4, 6},
}

// Game is a full game state. It is a plain value: copying it clones the game.
type Game struct {
	cells   [81]Player
	winners [9]Player // winner of each small board
	filled  [9]uint8  // occupied cells of each small board
	turn    Player
	forced  int8 // small board the next move must be played in, -1 for any open board
	last    Move
	winner  Player
	over    bool
}

// New starts a game where first plays the first move.
func New(first Player) Game {
	return Game{turn: first, forced: -1, last: NoMove}
}

// Turn returns the player who plays next.
func (g *Game) Turn() Player { return g.turn }

// Cell returns the owner of the cell played by m.
func (g *Game) Cell(m Move) Player { return g.cells[m] }

// LastMove returns the most recent move, or NoMove at the start of the game.
func (g *Game) LastMove() Move { return g.last }

// Over reports whether the game has ended.
func (g *Game) Over() bool { return g.over }

// Winner returns the winner of the game, Nobody while playing or on a draw.
func (g *Game) Winner() Player { return g.winner }

// BoardWinner returns the winner of a small board.
func (g *Game) BoardWinner(board int) Player { return g.winners[board] }

// BoardDone reports whether a small board is won or full.
func (g *Game) BoardDone(board int) bool {
	return g.winners[board] != Nobody || g.filled[board] == 9
}

// BoardPlayable reports whether the next move may be played in the given small board.
func (g *Game) BoardPlayable(board int) bool {
	if g.over || g.BoardDone(board) {
		return false
	}
	return g.forced < 0 || int(g.forced) == board
}

// Legal reports whether m can be played now.
func (g *Game) Legal(m Move) bool {
	return m >= 0 && m < 81 && g.cells[m] == Nobody && g.BoardPlayable(m.Board())
}

// LegalMoves appends every legal move to buf and returns it.
func (g *Game) LegalMoves(buf []Move) []Move {
	buf = buf[:0]
	for b := 0; b < 9; b++ {
		if !g.BoardPlayable(b) {
			continue
		}
		for m := MoveAt(b, 0); m < MoveAt(b+1, 0); m++ {
			if g.cells[m] == Nobody {
				buf = append(buf, m)
			}
		}
	}
	return buf
}

// Play applies m for the player whose turn it is. It reports false and leaves the game untouched if m is illegal.
func (g *Game) Play(m Move) bool {
	if !g.Legal(m) {
		return false
	}
	board := m.Board()
	g.cells[m] = g.turn
	g.filled[board]++
	if completesLine(g.smallBoard(board), m.Cell(), g.turn) {
		g.winners[board] = g.turn
		if completesLine(g.winners[:], board, g.turn) {
			g.winner, g.over = g.turn, true
		}
	}
	if !g.over && g.allBoardsDone() {
		g.over = true
	}
	g.last = m
	g.forced = int8(m.Cell())
	if g.BoardDone(m.Cell()) {
		g.forced = -1
	}
	g.turn = g.turn.Opponent()
	return true
}

func (g *Game) allBoardsDone() bool {
	for b := 0; b < 9; b++ {
		if !g.BoardDone(b) {
			return false
		}
	}
	return true
}

func (g *Game) smallBoard(board int) []Player { return g.cells[board*9 : board*9+9] }

// BoardLine returns the three cells that won a small board.
func (g *Game) BoardLine(board int) ([3]int, bool) {
	return findLine(g.smallBoard(board), g.winners[board])
}

// WinningLine returns the three small boards that won the game.
func (g *Game) WinningLine() ([3]int, bool) {
	return findLine(g.winners[:], g.winner)
}

// linesThrough lists, for each cell of a 3x3 grid, the lines that contain it.
var linesThrough = func() (through [9][][3]int) {
	for _, l := range Lines {
		for _, i := range l {
			through[i] = append(through[i], l)
		}
	}
	return through
}()

// completesLine reports whether p owns a full line of grid going through cell at.
func completesLine(grid []Player, at int, p Player) bool {
	for _, l := range linesThrough[at] {
		if grid[l[0]] == p && grid[l[1]] == p && grid[l[2]] == p {
			return true
		}
	}
	return false
}

func findLine(grid []Player, p Player) ([3]int, bool) {
	if p == Nobody {
		return [3]int{}, false
	}
	for _, l := range Lines {
		if grid[l[0]] == p && grid[l[1]] == p && grid[l[2]] == p {
			return l, true
		}
	}
	return [3]int{}, false
}

// MarshalText encodes the game as 81 cells (".", "O" or "X"), the player to move and the last move.
func (g *Game) MarshalText() ([]byte, error) {
	b := make([]byte, 0, 86)
	for _, c := range g.cells {
		b = append(b, cellChar(c))
	}
	b = append(b, cellChar(g.turn))
	return fmt.Appendf(b, "%d", g.last), nil
}

// UnmarshalText decodes a game encoded by MarshalText.
func (g *Game) UnmarshalText(text []byte) error {
	if len(text) < 83 {
		return fmt.Errorf("engine: game encoding too short (%d bytes)", len(text))
	}
	var last int
	if _, err := fmt.Sscanf(string(text[82:]), "%d", &last); err != nil || last < int(NoMove) || last >= 81 {
		return fmt.Errorf("engine: invalid last move %q", text[82:])
	}
	turn, err := parseCell(text[81])
	if err != nil || turn == Nobody {
		return fmt.Errorf("engine: invalid player to move %q", text[81])
	}

	*g = New(turn)
	for i := range g.cells {
		if g.cells[i], err = parseCell(text[i]); err != nil {
			return err
		}
		if g.cells[i] != Nobody {
			g.filled[i/9]++
		}
	}
	for _, p := range []Player{O, X} {
		for b := 0; b < 9; b++ {
			if _, ok := findLine(g.smallBoard(b), p); ok {
				g.winners[b] = p
			}
		}
	}
	for _, p := range []Player{O, X} {
		if _, ok := findLine(g.winners[:], p); ok {
			g.winner, g.over = p, true
		}
	}
	g.over = g.over || g.allBoardsDone()
	g.last = Move(last)
	if g.last != NoMove && !g.BoardDone(g.last.Cell()) {
		g.forced = int8(g.last.Cell())
	}
	return nil
}

func cellChar(p Player) byte {
	switch p {
	case O:
		return 'O'
	case X:
		return 'X'
	}
	return '.'
}

func parseCell(c byte) (Player, error) {
	switch c {
	case '.':
		return Nobody, nil
	case 'O':
		return O, nil
	case 'X':
		return X, nil
	}
	return Nobody, fmt.Errorf("engine: invalid cell %q", c)
}
