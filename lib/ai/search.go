//go:build !js

package ai

import (
	"runtime"
	"sync"
	"time"

	"GoTicTacToe/lib/engine"
)

// Search thinks about state for the given duration on every CPU core and returns the best move found.
func Search(state engine.Game, budget time.Duration) Result {
	deadline := time.Now().Add(budget)
	trees := make([][]MoveStat, runtime.NumCPU())
	var wg sync.WaitGroup
	for i := range trees {
		wg.Go(func() { trees[i] = Explore(state, deadline) })
	}
	wg.Wait()
	return Merge(trees...)
}
