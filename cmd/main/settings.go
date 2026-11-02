package main

import (
	"time"

	"GoTicTacToe/lib/ai"
)

// level describes how hard the computer plays.
type level struct {
	name   string
	about  string
	limits ai.Limits
	ponder bool // keep thinking during the player's turn
}

var levels = [...]level{
	{"Beginner", "Plays after 300 simulations", ai.Limits{Duration: 5 * time.Second, Simulations: 300}, false},
	{"Easy", "Plays after 3,000 simulations", ai.Limits{Duration: 5 * time.Second, Simulations: 3_000}, false},
	{"Medium", "Plays after 30,000 simulations", ai.Limits{Duration: 5 * time.Second, Simulations: 30_000}, false},
	{"Hard", "Thinks 2 s per move and during your turn", ai.Limits{Duration: 2 * time.Second}, true},
	{"Expert", "Thinks 6 s per move and during your turn", ai.Limits{Duration: 6 * time.Second}, true},
}

// firstMove chooses who starts each game.
type firstMove int

const (
	firstAlternate firstMove = iota
	firstO
	firstX
)

// settings are the preferences and score kept between sessions.
type settings struct {
	VsComputer bool      `json:"vsComputer"`
	Level      int       `json:"level"`   // index in levels
	Threads    int       `json:"threads"` // threads the computer uses, 0 for all
	FirstMove  firstMove `json:"firstMove"`
	Sound      bool      `json:"sound"`
	Dark       bool      `json:"dark"`
	SeenRules  bool      `json:"seenRules"`
	Score      [3]int    `json:"score"` // wins indexed by player, draws under engine.Nobody
}

func defaultSettings() settings {
	return settings{VsComputer: true, Level: 2, Sound: true, Dark: prefersDark()}
}

// loadSettings returns the saved settings, or the defaults when there are none or they are invalid.
func loadSettings() settings {
	s, ok := readSettings()
	if !ok || s.Level < 0 || s.Level >= len(levels) || s.Threads < 0 || s.FirstMove < firstAlternate || s.FirstMove > firstX {
		return defaultSettings()
	}
	return s
}
