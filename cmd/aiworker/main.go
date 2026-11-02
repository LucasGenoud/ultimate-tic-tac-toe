//go:build js

// Command aiworker runs inside a Web Worker and serves tree searches to the game page (see wasm/ai-worker.js).
package main

import (
	"encoding/json"
	"syscall/js"
	"time"

	"GoTicTacToe/lib/ai"
	"GoTicTacToe/lib/engine"
)

// search(state string, milliseconds number) string returns the JSON encoded root statistics.
func search(_ js.Value, args []js.Value) any {
	var state engine.Game
	if err := state.UnmarshalText([]byte(args[0].String())); err != nil {
		return js.Global().Get("Error").New(err.Error())
	}
	deadline := time.Now().Add(time.Duration(args[1].Int()) * time.Millisecond)
	stats, err := json.Marshal(ai.Explore(state, deadline))
	if err != nil {
		return js.Global().Get("Error").New(err.Error())
	}
	return string(stats)
}

func main() {
	js.Global().Call("onSearchReady", js.FuncOf(search))
	select {}
}
