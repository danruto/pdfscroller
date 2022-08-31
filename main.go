package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	log.Println("[Core] Started PDFScroller")

	args := os.Args[1:]
	log.Println(fmt.Sprintf("[Core] Started with arguments: %s", args))

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("PDFScroller")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g, err := NewGame(args[0])
	if err != nil {
		log.Fatal(err)
	}

	g.CacheImages()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

	log.Println("[Core] Done")
}