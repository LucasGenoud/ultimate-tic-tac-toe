# Ultimate tic-tac-toe


## Description
GO implementation of an ultimate tic-tac-toe with a monte carlo search tree AI

Realised during the course "Programmation élégante en GO" from HEIA-FR. Initially forked, refactored and improved upon this project: https://github.com/LempekPL/GoTicTacToe

## Playing

- Click a cell inside a highlighted board. Your move decides which board your opponent plays in next.
- Win three small boards in a row to win the game.
- Play against the computer (levels 1 to 5 set how many seconds it thinks per move) or against a friend.
- Hold `R` to reset the score, hold `Esc` to quit (desktop only).

## Building

Requires Go 1.25 or later.

```bash
make test    # rules and AI tests
make build   # desktop binary
make serve   # web build served on http://localhost:8080
```

## How the AI works

The computer runs a Monte Carlo tree search on every CPU core at once (root parallelisation):
each core grows its own tree for the allotted time, then the statistics of the candidate moves are merged.
On the desktop each tree runs in a goroutine. In the browser, where WebAssembly runs Go on a single thread,
the trees run in one Web Worker per core (`cmd/aiworker`, `wasm/ai-pool.js`), which also keeps the page responsive while it thinks.

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
