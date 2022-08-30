package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	log.Println("Started PDFScroller")

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("PDFScroller")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

	log.Println("Done")
}