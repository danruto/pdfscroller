package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"math"
	"os"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/rs/zerolog/log"
)

const (
	screenWidth  = 800
	screenHeight = 1200
)

// NewImageFromReader creates a new image from a data stream.
func NewImageFromReader(read io.Reader) *image.Image {
	data, err := ioutil.ReadAll(read)
	if err != nil {
		log.Error().Err(err).Str("function", "NewImageFromReader").Msg("Unable to read image")
		return nil
	}

	ret, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Error().Err(err).Str("function", "NewImageFromReader").Msg("Unable to read image")
		return nil
	}

	return &ret
}

type GamePage struct {
	position int
	y        float64
}

func (page *GamePage) draw(g *Game, screen *ebiten.Image) {
	if page != nil {
		width, _ := g.images[page.position].Size()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(g.zoom, g.zoom)
		op.GeoM.Translate(float64(screenWidth/2-(float64(width)*g.zoom)/2), page.y*g.zoom)
		screen.DrawImage(g.images[page.position], op)
	}
}

type Game struct {
	file      *os.File
	maxImages int
	images    []*ebiten.Image
	speed     float64
	zoom      float64

	// Window of pages to render, only a maximum of 2
	pages [2]*GamePage

	// DEBUG
	update bool
}

func cacheImages(f *os.File, pageSelections []string, offset int64) ([]*ebiten.Image, error) {
	images, err := pdfapi.ExtractImagesRaw(f, pageSelections, nil)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("function", "cacheImages").Msg(fmt.Sprintf("found %d images", len(images)))

	var gameImages = make([]*ebiten.Image, len(pageSelections))
	for _, image := range images {
		v := reflect.ValueOf(image).FieldByName("pageNr")
		img := NewImageFromReader(image.Reader)

		gameImages[v.Int()-1-offset] = ebiten.NewImageFromImage(*img)

		log.Debug().
			Str("function", "cacheImages").
			Int64("pageNr", v.Int()).
			Int64("offset", offset).
			Msg(fmt.Sprintf("Inserting into position: %d", v.Int()-1-offset))
	}

	log.Debug().Str("function", "cacheImages").Msg(fmt.Sprintf("appended %d images", len(gameImages)))
	return gameImages, nil
}

func (g *Game) CacheImages() {
	offset := 2
	remainingSize := g.maxImages - offset
	if remainingSize < 0 {
		remainingSize = 0
	}

	go func(g *Game, remainingSize int) error {
		var pageSelections = make([]string, remainingSize)
		ii := 0
		for ii < remainingSize {
			// Skip 2 which we already know we have fetched and then add 1 more to make it a 1-based index
			pageSelections[ii] = fmt.Sprintf("%d", ii+offset+1)
			ii += 1
		}

		log.Debug().Str("function", "CacheImages").Msg(fmt.Sprintf("Skipped: %d, Remaining Size: %d, Page selections: %v", offset, remainingSize, pageSelections))

		images, err := cacheImages(g.file, pageSelections, int64(offset))
		if err != nil {
			return err
		}
		// TODO: We can chunk this further but for now it's fine. By the time we read 2 pages,
		// it should have been enough time to load the rest
		g.images = append(g.images, images...)
		return nil
	}(g, remainingSize)
}

func NewGame(filename string) (*Game, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	imageCount, err := pdfapi.PageCount(f, nil)
	if err != nil {
		return nil, err
	}
	log.Debug().Str("function", "NewGame").Msg(fmt.Sprintf("PageCount: %d", imageCount))

	images, err := cacheImages(f, []string{"1", "2"}, 0)
	if err != nil {
		return nil, err
	}

	g := &Game{
		file:      f,
		maxImages: imageCount,
		images:    images,
		speed:     0.0,
		zoom:      1.0,
		pages: [2]*GamePage{
			{
				position: 0,
				y:        0,
			},
			nil,
		},

		update: true,
	}

	return g, nil
}

func (g *Game) handleKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		g.speed -= 1.0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		g.speed += 1.0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.speed -= 40.0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.speed += 40.0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.speed = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		// Find the position of previous
		newPosition := g.pages[0].position - 1
		if newPosition >= 0 {
			log.
				Debug().
				Str("function", "handleKeys").
				Int("prevPosition", g.pages[0].position).
				Int("newPosition", newPosition).
				Int("maxImages", g.maxImages).
				Msg("Previous page requested")

			g.pages[0] = &GamePage{
				y:        0,
				position: newPosition,
			}
			g.pages[1] = nil
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		// Find the position of next
		newPosition := 0

		if g.pages[1] == nil {
			newPosition = g.pages[0].position + 1
		} else {
			newPosition = g.pages[1].position + 1
		}

		if newPosition < g.maxImages {
			log.
				Debug().
				Str("function", "handleKeys").
				Int("prevPosition", g.pages[0].position).
				Int("newPosition", newPosition).
				Int("maxImages", g.maxImages).
				Msg("Next page requested")

			g.pages[0] = &GamePage{
				y:        0,
				position: newPosition,
			}
			g.pages[1] = nil
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		g.zoom += 0.1
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.zoom -= 0.1
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.update = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.update = false
	}
}

func (g *Game) handleMouseInputs() {
}

func (g *Game) Update() error {

	// Handle any key presses
	g.handleKeys()

	// Handle any mouse actions
	g.handleMouseInputs()

	// Don't update if our speed is 0
	if g.speed == 0 {
		return nil
	}

	pageOne := g.pages[0]
	pageTwo := g.pages[1]

	// TODO: Handle g.images when we are still loading
	// TODO: Handle zoom calculations

	// We should always have a pageOne, if we don't panic
	if pageOne == nil {
		panic("Page One is somehow empty. Unrecoverable error.")
	}

	// Setup specific page values
	pageOneHeight := float64(g.images[pageOne.position].Bounds().Size().Y)
	pageOneY := pageOne.y
	pageOneYAbs := math.Abs(pageOneY)
	pageTwoHeight := 0.
	pageTwoY := 0.
	if pageTwo != nil {
		pageTwoHeight = float64(g.images[pageTwo.position].Bounds().Size().Y)
		pageTwoY = pageTwo.y
	}

	// Setup some common values to use for our calculations
	scroll := true

	if g.speed > 0 {
		if pageOneYAbs >= pageOneHeight {
			// Shift pageTwo into pageOne when we have finished reading pageOne
			g.pages[0] = pageTwo
			g.pages[1] = nil

			log.
				Debug().
				Str("function", "Update").
				Float64("speed", g.speed).
				Float64("pageOneYAbs", pageOneYAbs).
				Float64("pageOneHeight", pageOneHeight).
				Msg("Shifted pageTwo into pageOne")
		} else if pageOneYAbs >= float64(pageOneHeight-screenHeight) {
			// Pre-load the next image if we are getting close to it which we define as one 'screenHeight' away, assuming pageHeight > screenHeight
			nextPosition := pageOne.position + 1

			// Only try and continue if the next position is valid and we haven't already set a next page
			// We shift out pageTwo in the above condition so we know we can gate it by nil
			if pageTwo == nil && nextPosition < g.maxImages {
				// Start the new page just underneath the current page i.e.
				// at +screenHeight
				delta := 0.
				if g.zoom != 1.0 {
					delta = (1 - g.zoom) * screenHeight
				}
				// TODO: Account for scroll velocity
				g.pages[1] = &GamePage{
					// y:        float64(g.images[nextPosition].Bounds().Size().Y),
					y:        screenHeight + delta,
					position: nextPosition,
				}

				log.
					Debug().
					Str("function", "Update").
					Float64("speed", g.speed).
					Int("nextPosition", nextPosition).
					Int("maxImages", g.maxImages).
					Float64("pageOneYAbs", pageOneYAbs).
					Float64("bounds", float64(pageOneHeight-screenHeight)).
					Float64("delta", delta).
					Msg("Appending a new page to the bottom as we have hit the boundary")
			} else if nextPosition == g.maxImages && pageOneYAbs+screenHeight >= float64(pageOneHeight) {
				scroll = false

				log.
					Debug().
					Str("function", "Update").
					Float64("speed", g.speed).
					Int("nextPosition", nextPosition).
					Int("maxImages", g.maxImages).
					Float64("pageOneYAbs", pageOneYAbs).
					Int("screenHeight", screenHeight).
					Float64("pageOneHeight", pageOneHeight).
					Msg("Reached the end of the pdf")
			}
		}

	} else {
		prevPosition := pageOne.position - 1

		if pageTwo != nil && pageTwoY > screenHeight {
			g.pages[1] = nil

			log.Debug().
				Str("function", "Update").
				Float64("speed", g.speed).
				Float64("pageOneY", pageOneY).
				Float64("pageTwoY", pageTwoY).
				Float64("pageTwoHeight", pageTwoHeight).
				Int("screenHeight", screenHeight).
				Msg("PageTwo is offscreen, removing it")
		} else if (pageOneYAbs == 0 || pageOneY >= 0) && prevPosition >= 0 {
			g.pages[1] = pageOne
			g.pages[1].y = 0

			delta := 0.
			if g.zoom != 1.0 {
				delta = (1 - g.zoom) * screenHeight
			}
			// TODO: Account for scroll velocity
			g.pages[0] = &GamePage{
				y:        -float64(g.images[prevPosition].Bounds().Size().Y) - delta,
				position: prevPosition,
			}

			log.Debug().
				Str("function", "Update").
				Float64("speed", g.speed).
				Float64("pageOneY", pageOneY).
				Int("prevPosition", prevPosition).
				Float64("delta", delta).
				Msg("Shifting pageOne to pageTwo as we have hit the top of the page")
		} else if (pageOneYAbs == 0 || pageOneY > 0) && prevPosition < 0 && pageTwo == nil {
			// We have reached the top of the first page
			// Clamp to 0 in case of overscroll
			originalPageOneY := pageOneY
			g.pages[0].y = 0

			// Disable scrolling
			scroll = false

			log.Debug().
				Str("function", "Update").
				Float64("speed", g.speed).
				Float64("pageOneYAbs", pageOneYAbs).
				Float64("pageOneY", originalPageOneY).
				Bool("pageTwo exists?", pageTwo != nil).
				Int("prevPosition", prevPosition).
				Msg("We have reached the top of the first page. Clamping and disabling scroll.")
		}
	}

	if scroll {
		pageOne.y -= g.speed
		if pageTwo != nil {
			pageTwo.y -= g.speed
		}
	}

	return nil
}

func (g *Game) drawDebug(screen *ebiten.Image) {
	position := 0
	// Positive speed means we take the highest position
	// Negative means we take the lowest position
	if g.speed >= 0 {
		for _, page := range g.pages {
			if page != nil {
				position = page.position + 1
			}
		}
	} else {
		position = g.pages[0].position + 1
	}

	pageTwoY := 0.
	pageTwo := g.pages[1]
	if pageTwo != nil {
		pageTwoY = pageTwo.y
	}

	ebitenutil.DebugPrint(
		screen,
		fmt.Sprintf("Page: %d of %d (Loaded: %d), Speed: %0.2f, Zoom: %0.2f, PageOne: %0.2f, PageTwo: %0.2f, TPS: %0.2f",
			position,
			g.maxImages,
			len(g.images),
			g.speed,
			g.zoom,
			g.pages[0].y,
			pageTwoY,
			ebiten.CurrentTPS(),
		),
	)
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, page := range g.pages {
		page.draw(g, screen)
	}

	g.drawDebug(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	log.
		Info().
		Str("function", "Layout").
		Int("outsideWidth", outsideWidth).
		Int("outsideHeight", outsideHeight).
		Int("screenWidth", screenWidth).
		Int("screenHeight", screenHeight).
		Msg("Called layout")
	return screenWidth, screenHeight
}
