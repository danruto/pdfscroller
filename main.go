package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// os.Setenv("EBITENGINE_GRAPHICS_LIBRARY", "opengl")
	// zerolog.SetGlobalLevel(zerolog.TraceLevel)
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	log.Trace().Str("function", "main").Msg("Started PDFScroller")

	args := os.Args[1:]
	log.Debug().Msg(fmt.Sprintf("[Core] Started with arguments: %s", args))

	if len(args) != 1 {
		panic("Only 1 argument is required. This is the path to the pdf file.")
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(fmt.Sprintf("PDFScroller - %s", args[0]))
	// ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

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