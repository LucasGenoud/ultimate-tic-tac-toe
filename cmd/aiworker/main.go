//go:build js

// Command aiworker runs inside a Web Worker and grows a search tree for the game page
// (see wasm/ai-worker.js). The tree is kept between calls, so positions reached from
// the previous one reuse it.
package main

import (
	"encoding/json"
	"syscall/js"
	"time"

	"GoTicTacToe/lib/ai"
	"GoTicTacToe/lib/engine"
)

// step(state string, milliseconds number, simulations number, untilFull bool) {played: number, stats: string}
// searches state for at most the given time and playouts (0 for no limit), or until the tree
// is full when untilFull is set, and returns the playouts run and the JSON encoded root statistics.
func step(worker *ai.Worker) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) any {
		var state engine.Game
		if err := state.UnmarshalText([]byte(args[0].String())); err != nil {
			return js.Global().Get("Error").New(err.Error())
		}
		stats, played := worker.Search(state, ai.Budget{
			Deadline:    time.Now().Add(time.Duration(args[1].Int()) * time.Millisecond),
			Simulations: args[2].Int(),
			UntilFull:   args[3].Bool(),
		})
		encoded, err := json.Marshal(stats)
		if err != nil {
			return js.Global().Get("Error").New(err.Error())
		}
		return map[string]any{"played": played, "stats": string(encoded)}
	})
}

func main() {
	threads := js.Global().Get("poolSize").Int() // set by ai-worker.js
	js.Global().Call("onStepReady", step(ai.NewWorker(ai.NodesPerThread(threads))))
	select {}
}
