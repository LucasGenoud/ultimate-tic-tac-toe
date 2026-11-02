# Ultimate tic-tac-toe


## Description
GO implementation of an ultimate tic-tac-toe with a monte carlo search tree AI

Realised during the course "Programmation élégante en GO" from HEIA-FR. Initially forked, refactored and improved upon this project: https://github.com/LempekPL/GoTicTacToe

## Playing

- Click a cell inside a highlighted board. Your move decides which board your opponent plays in next.
- Win three small boards in a row to win the game. "How to play" in the menu explains the rules.
- Play against the computer, from Beginner to Expert, or against a friend, and choose who moves first.
- While playing: `H` or Hint suggests a move, `U`, `Ctrl+Z` or Undo takes your last move back, `N` opens the menu.
- Anywhere: `M` toggles sound, `D` toggles the dark theme, hold `R` to reset the score, hold `Esc` to quit (desktop only).
- Settings and score are saved: in the user configuration directory on the desktop, in local storage on the web.

## Building

Requires Go 1.25 or later.

```bash
make test    # rules and AI tests
make build   # desktop binary
make serve   # web build served on http://localhost:8080
```

## How the AI works

The computer runs a Monte Carlo tree search on several threads at once (root parallelisation):
each thread grows its own tree, then the statistics of the candidate moves are merged.
The number of threads can be set in the menu. On the desktop each tree runs in a goroutine. In the browser,
where WebAssembly runs Go on a single thread, the trees run in Web Workers (`cmd/aiworker`, `wasm/ai-pool.js`),
which also keeps the page responsive; without Web Workers the search runs on the page, on one thread.

- Beginner to Medium stop after a fixed number of simulations; Hard and Expert think for a fixed time.
- Trees are kept from one move to the next, and Hard and Expert keep thinking during your turn.
- Positions won or lost by force are proven, so the computer never misses a winning move and shows a forced result.
- The statistics of the search (simulations, the computer's estimated chance of winning) update live below the board.

## Project layout

- `lib/engine`: game rules, independent of rendering
- `lib/ai`: tree search and its native and browser parallel runners
- `lib/graphics`: resolution independent drawing helpers
- `cmd/main`: the game (ebiten)
- `cmd/aiworker`, `wasm/`: browser AI worker and web page

## License
Copyright 2024 Lucas Genoud, Benoît Vorlet

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
