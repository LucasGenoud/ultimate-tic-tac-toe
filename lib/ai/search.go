//go:build !js

package ai

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"GoTicTacToe/lib/engine"
)

// ponderLimit bounds how long the searcher thinks during the opponent's turn.
const ponderLimit = 60 * time.Second

// Searcher runs one tree per CPU core in goroutines.
type Searcher struct {
	workers []*Worker
	lock    sync.Mutex                  // serialises searches and pondering
	running sync.WaitGroup              // goroutines of the current search or pondering
	stop    atomic.Pointer[atomic.Bool] // stop flag of the current search or pondering
	cancels atomic.Int64                // incremented by Cancel, so a search queued before it does not start

	progressMu sync.Mutex
	progress   [][]MoveStat // latest root statistics of each thread of the current search
}

// NewSearcher returns a searcher able to use every CPU core.
func NewSearcher() *Searcher {
	n := runtime.NumCPU()
	s := &Searcher{workers: make([]*Worker, n)}
	s.stop.Store(new(atomic.Bool))
	for i := range s.workers {
		s.workers[i] = NewWorker(NodesPerThread(n))
	}
	return s
}

// MaxThreads returns how many threads the searcher can use.
func (s *Searcher) MaxThreads() int { return len(s.workers) }

// Fallback reports whether the searcher lost its threads; it never does natively.
func (s *Searcher) Fallback() bool { return false }

// Search thinks about state within the limits and returns the best move found. It blocks,
// stops any pondering first, and returns early with the best move so far when Cancel is called.
func (s *Searcher) Search(state engine.Game, limits Limits) Result {
	cancels := s.cancels.Load()
	s.lock.Lock()
	defer s.lock.Unlock()
	stop, ok := s.restart(cancels)
	if !ok {
		return Result{Move: engine.NoMove}
	}

	workers := s.threads(limits.Threads)
	s.resetProgress(len(workers))
	deadline := time.Now().Add(limits.Duration)
	perThread := 0
	if limits.Simulations > 0 {
		perThread = (limits.Simulations + len(workers) - 1) / len(workers)
	}
	for i, w := range workers {
		s.running.Go(func() {
			stats, _ := w.Search(state, Budget{
				Deadline:    deadline,
				Simulations: perThread,
				Stop:        stop,
				Progress:    func(stats []MoveStat) { s.setProgress(i, stats) },
			})
			s.setProgress(i, stats)
		})
	}
	s.running.Wait()
	return s.Progress()
}

// Ponder keeps growing the trees of state in the background until the next Search or Cancel.
func (s *Searcher) Ponder(state engine.Game, threads int) {
	cancels := s.cancels.Load()
	s.lock.Lock()
	defer s.lock.Unlock()
	stop, ok := s.restart(cancels)
	if !ok {
		return
	}
	deadline := time.Now().Add(ponderLimit)
	for _, w := range s.threads(threads) {
		s.running.Go(func() {
			w.Search(state, Budget{Deadline: deadline, UntilFull: true, Stop: stop})
		})
	}
}

// Cancel stops the running search or pondering without waiting.
func (s *Searcher) Cancel() {
	s.cancels.Add(1)
	s.stop.Load().Store(true)
}

// Progress returns the merged statistics of the current or last search.
func (s *Searcher) Progress() Result {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()
	return Merge(s.progress...)
}

// restart stops the previous search or pondering, waits for it, and returns the stop flag of
// the next one. It reports false when Cancel was called since cancels was read.
func (s *Searcher) restart(cancels int64) (*atomic.Bool, bool) {
	s.stop.Load().Store(true)
	s.running.Wait()
	stop := new(atomic.Bool)
	s.stop.Store(stop)
	return stop, s.cancels.Load() == cancels
}

func (s *Searcher) threads(n int) []*Worker {
	if n <= 0 || n > len(s.workers) {
		n = len(s.workers)
	}
	return s.workers[:n]
}

func (s *Searcher) resetProgress(threads int) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()
	s.progress = make([][]MoveStat, threads)
}

func (s *Searcher) setProgress(i int, stats []MoveStat) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()
	s.progress[i] = stats
}
