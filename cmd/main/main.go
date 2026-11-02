// Command main runs the ultimate tic-tac-toe game, natively or in the browser.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(640, 800)
	ebiten.SetWindowTitle("Ultimate Tic-Tac-Toe")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewApp()); err != nil {
		log.Fatal(err)
	}
}
