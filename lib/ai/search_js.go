//go:build js

package ai

import (
	"encoding/json"
	"syscall/js"
	"time"

	"GoTicTacToe/lib/engine"
)

// fallbackSlice is how long the in-page fallback searches before letting the browser render a frame.
const fallbackSlice = 15 * time.Millisecond

// Search thinks about state for the given duration and returns the best move found.
//
// WebAssembly runs Go on a single thread, so the search is spread over the Web Workers
// started by wasm/ai-pool.js (one per core). Without them it runs in the page itself.
func Search(state engine.Game, budget time.Duration) Result {
	if pool := js.Global().Get("aiPool"); pool.Truthy() {
		if res, ok := searchInWorkers(pool, state, budget); ok {
			return res
		}
	}
	return searchInPage(state, budget)
}

// searchInWorkers asks every worker of the pool for its root statistics and merges them.
func searchInWorkers(pool js.Value, state engine.Game, budget time.Duration) (Result, bool) {
	encoded, err := state.MarshalText()
	if err != nil {
		return Result{}, false
	}

	replies := make(chan []js.Value, 1)
	onDone := js.FuncOf(func(_ js.Value, args []js.Value) any {
		replies <- args
		return nil
	})
	defer onDone.Release()
	pool.Call("search", string(encoded), budget.Milliseconds()).Call("then", onDone, onDone)

	reply := <-replies
	if len(reply) == 0 || !js.Global().Get("Array").Call("isArray", reply[0]).Bool() {
		return Result{}, false // the promise was rejected
	}
	trees := make([][]MoveStat, reply[0].Length())
	for i := range trees {
		if err := json.Unmarshal([]byte(reply[0].Index(i).String()), &trees[i]); err != nil {
			return Result{}, false
		}
	}
	return Merge(trees...), len(trees) > 0
}

// searchInPage runs a single tree in short slices so the page keeps rendering while it thinks.
func searchInPage(state engine.Game, budget time.Duration) Result {
	t := newTree(state)
	deadline := time.Now().Add(budget)
	for time.Now().Before(deadline) {
		slice := time.Now().Add(fallbackSlice)
		if slice.After(deadline) {
			slice = deadline
		}
		t.run(slice)
		time.Sleep(time.Millisecond) // hands control back to the browser event loop
	}
	return Merge(t.stats())
}
