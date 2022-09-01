package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Trace().Str("function", "main").Msg("Started PDFScroller")

	args := os.Args[1:]
	log.Debug().Msg(fmt.Sprintf("[Core] Started with arguments: %s", args))

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("PDFScroller")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g, err := NewGame(args[0])
	if err != nil {
		log.Error().Err(err)
	}

	g.CacheImages()

	if err := ebiten.RunGame(g); err != nil {
		log.Error().Err(err)
	}

	log.Trace().Str("function", "main").Msg("Done")
}