//go:build js

package ai

import (
	"encoding/json"
	"sync/atomic"
	"syscall/js"
	"time"

	"GoTicTacToe/lib/engine"
)

const (
	// ponderLimit bounds how long the searcher thinks during the opponent's turn.
	ponderLimit = 60 * time.Second
	// fallbackSlice is how long the in-page fallback searches before letting the browser render a frame.
	fallbackSlice = 15 * time.Millisecond
)

// Searcher spreads the search over the Web Workers started by wasm/ai-pool.js, one per core,
// since WebAssembly runs Go on a single thread. Without them it searches on the page itself.
type Searcher struct {
	local    *Worker // in-page fallback
	stop     atomic.Bool
	cancels  atomic.Int64 // incremented by Cancel, so a search queued before it does not start
	progress Result       // progress of the in-page fallback
}

func NewSearcher() *Searcher {
	return &Searcher{local: NewWorker(NodesPerThread(1))}
}

// pool returns the worker pool, which is removed if one of its workers fails.
func pool() js.Value { return js.Global().Get("aiPool") }

// MaxThreads returns how many threads the searcher can use.
func (s *Searcher) MaxThreads() int {
	if p := pool(); p.Truthy() {
		return p.Get("size").Int()
	}
	return 1
}

// Fallback reports whether Web Workers are unavailable, the search then running on the page.
func (s *Searcher) Fallback() bool { return !pool().Truthy() }

// Search thinks about state within the limits and returns the best move found. It blocks the
// calling goroutine only, and returns early with the best move so far when Cancel is called.
func (s *Searcher) Search(state engine.Game, limits Limits) Result {
	cancels := s.cancels.Load()
	if p := pool(); p.Truthy() {
		if res, ok := s.searchInWorkers(p, state, limits); ok {
			return res
		}
	}
	return s.searchInPage(state, limits, cancels)
}

// Ponder keeps growing the trees of state in the workers until the next Search or Cancel.
// The in-page fallback does not ponder, to keep the page responsive.
func (s *Searcher) Ponder(state engine.Game, threads int) {
	if p := pool(); p.Truthy() {
		p.Call("ponder", encode(state), ponderLimit.Milliseconds(), threads)
	}
}

// Cancel stops the running search or pondering without waiting.
func (s *Searcher) Cancel() {
	s.cancels.Add(1)
	s.stop.Store(true)
	if p := pool(); p.Truthy() {
		p.Call("stop")
	}
}

// Progress returns the merged statistics of the current or last search.
func (s *Searcher) Progress() Result {
	p := pool()
	if !p.Truthy() {
		return s.progress
	}
	trees, ok := decodeStats(p.Call("progress"))
	if !ok {
		return Result{Move: engine.NoMove}
	}
	return Merge(trees...)
}

func (s *Searcher) searchInWorkers(p js.Value, state engine.Game, limits Limits) (Result, bool) {
	replies := make(chan js.Value, 1)
	onDone := js.FuncOf(func(_ js.Value, args []js.Value) any {
		replies <- args[0]
		return nil
	})
	defer onDone.Release()
	p.Call("search", encode(state), limits.Duration.Milliseconds(), limits.Simulations, limits.Threads).
		Call("then", onDone, onDone)

	trees, ok := decodeStats(<-replies) // a rejected promise passes an Error instead of an array
	if !ok {
		return Result{}, false
	}
	return Merge(trees...), true
}

// searchInPage runs a single tree in short slices so the page keeps rendering while it thinks.
func (s *Searcher) searchInPage(state engine.Game, limits Limits, cancels int64) Result {
	s.stop.Store(false)
	deadline := time.Now().Add(limits.Duration)
	for played := 0; time.Now().Before(deadline) && !s.stop.Load() && s.cancels.Load() == cancels; {
		remaining := 0
		if limits.Simulations > 0 {
			if remaining = limits.Simulations - played; remaining <= 0 {
				break
			}
		}
		slice := time.Now().Add(fallbackSlice)
		if slice.After(deadline) {
			slice = deadline
		}
		stats, n := s.local.Search(state, Budget{Deadline: slice, Simulations: remaining})
		played += n
		s.progress = Merge(stats)
		if n == 0 {
			break // the position is decided
		}
		time.Sleep(time.Millisecond) // hands control back to the browser event loop
	}
	return s.progress
}

func encode(state engine.Game) string {
	text, _ := state.MarshalText()
	return string(text)
}

// decodeStats parses an array of JSON encoded root statistics, one per worker.
func decodeStats(v js.Value) ([][]MoveStat, bool) {
	if !js.Global().Get("Array").Call("isArray", v).Bool() {
		return nil, false
	}
	trees := make([][]MoveStat, v.Length())
	for i := range trees {
		if err := json.Unmarshal([]byte(v.Index(i).String()), &trees[i]); err != nil {
			return nil, false
		}
	}
	return trees, true
}
