package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
)

const (
	screenWidth  = 800
	screenHeight = 1200
)

// NewImageFromReader creates a new image from a data stream.
func NewImageFromReader(read io.Reader) *image.Image {
	data, err := ioutil.ReadAll(read)
	if err != nil {
		log.Println("Unable to read image", err)
		return nil
	}

	ret, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Println("Unable to read image", err)
		return nil
	}

	return &ret
}

type GamePage struct {
	position int
	y        float64
}

func (page *GamePage) update(g *Game) {
	if page != nil {
		// Y is inverted, so get the absolute value for our calculation
		absY := math.Abs(page.y)
		maxY := float64(g.images[page.position].Bounds().Size().Y)

		bounds := maxY - screenHeight

		scroll := true

		if g.speed > 0 {
			if absY >= maxY {
				// Shift pages left when we have finished reading the first page
				g.pages[0] = g.pages[1]
				g.pages[1] = nil
				log.Println(fmt.Sprintf("[Update] Shifting pages left because absY: %0.2f, maxY: %0.2f", absY, maxY))
			} else if absY >= bounds {
				// If our y is approximately at the border of the next image, we render 2 images
				if g.pages[1] == nil && page.position < g.maxImages-1 {
					g.pages[1] = &GamePage{
						y:        screenHeight,
						position: page.position + 1,
					}
					log.Println(fmt.Sprintf("[Update] Adding new page because: absY: %0.2f, maxY: %0.2f, bounds: %0.2f", absY, maxY, bounds))
				} else if page.position == len(g.images)-1 {
					// If we have finished reading the pdf, then stop scrolling
					scroll = false
				}
			}
		} else if g.speed < 0 {
			// log.Println(fmt.Sprintf("[Update] Speed: %0.2f, absY: %0.2f, page: %v", g.speed, absY, page))
			newPosition := page.position - 1
			if page.y == 0 && newPosition >= 0 {
				log.Println(fmt.Sprintf("[Update Start] Shifting pages right because absY: %0.2f, pageY: %0.2f, p0: %v, p1: %v", absY, page.y, g.pages[0], g.pages[1]))
				// Shift pages right when we are scrolling backwards and have hit the top of the page
				g.pages[1] = g.pages[0]
				g.pages[1].y = 0

				g.pages[0] = &GamePage{
					y:        -float64(g.images[newPosition].Bounds().Size().Y),
					position: newPosition,
				}
				log.Println(fmt.Sprintf("[Update End] Shifting pages right because absY: %0.2f, pageY: %0.2f, p0: %v, p1: %v", absY, page.y, g.pages[0], g.pages[1]))
			} else if page.y == 0 && newPosition < 0 {
				// We have reached the top of the first page
				scroll = false
			}
		} else {
			// If our speed is 0, we don't scroll
			scroll = false
		}

		if scroll {
			// if g.speed > 0 {
			// 	page.y -= g.speed
			// } else {
			// 	page.y += g.speed
			// }
			page.y -= g.speed
		}
	}
}

func (page *GamePage) draw(g *Game, screen *ebiten.Image) {
	if page != nil {
		width, _ := g.images[page.position].Size()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(screenWidth/2-width/2), page.y)
		screen.DrawImage(g.images[page.position], op)
	}
}

type Game struct {
	file      *os.File
	maxImages int
	images    []*ebiten.Image
	speed     float64

	// Window of pages to render, only a maximum of 2
	pages [2]*GamePage
}

func cacheImages(f *os.File, pageSelections []string) ([]*ebiten.Image, error) {
	images, err := pdfapi.ExtractImagesRaw(f, pageSelections, nil)
	if err != nil {
		return nil, err
	}
	log.Println(fmt.Sprintf("[ExtractImagesRaw]: found %d images", len(images)))

	var gameImages []*ebiten.Image
	for _, image := range images {
		img := NewImageFromReader(image.Reader)

		gameImages = append(gameImages, ebiten.NewImageFromImage(*img))
	}

	log.Println(fmt.Sprintf("[ExtractImagesRaw] appended %d images", len(gameImages)))
	return gameImages, nil
}

func (g *Game) CacheImages() {
	remainingSize := g.maxImages - 2
	if remainingSize < 0 {
		remainingSize = 0
	}

	go func(g *Game, remainingSize int) error {
		var pageSelections = make([]string, remainingSize)
		ii := 0
		for ii < remainingSize {
			// Skip 2 which we already know we have fetched and then add 1 more to make it a 1-based index
			pageSelections[ii] = fmt.Sprintf("%d", ii+2+1)
			ii += 1
		}
		log.Println(fmt.Sprintf("[Goroutine] Skipped: 2, Remaining Size: %d, Page selections: %v", remainingSize, pageSelections))
		images, err := cacheImages(g.file, pageSelections)
		if err != nil {
			return err
		}
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
	log.Println(fmt.Sprintf("[NewGame] PageCount: %d", imageCount))

	images, err := cacheImages(f, []string{"1", "2"})
	if err != nil {
		return nil, err
	}

	g := &Game{
		file:      f,
		maxImages: imageCount,
		images:    images,
		speed:     4.0,
		pages: [2]*GamePage{
			{
				position: 0,
				y:        0,
			},
			nil,
		},
	}

	return g, nil
}

func (g *Game) handleKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		g.speed -= 0.2
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		g.speed += 0.2
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		g.speed -= 1.0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		g.speed += 1.0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.speed = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		// Find the position of previous
		newPosition := g.pages[0].position - 1
		if newPosition >= 0 {
			log.Println(fmt.Sprintf("[Input] Previous page requested, P0: %d, New: %d", g.pages[0].position, newPosition))
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

		if newPosition < g.maxImages-1 {
			log.Println(fmt.Sprintf("[Input] Next page requested, P0: %d, P1: %v, New: %d", g.pages[0].position, g.pages[1], newPosition))
			g.pages[0] = &GamePage{
				y:        0,
				position: newPosition,
			}
			g.pages[1] = nil
		}
	}
}

func (g *Game) Update() error {
	for _, page := range g.pages {
		// log.Println(fmt.Sprintf("Page: %d, Nil?: %v, Page?: %v", ii, page == nil, page))
		page.update(g)
	}

	g.handleKeys()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Only render the image window based on the y and position
	for _, page := range g.pages {
		page.draw(g, screen)
	}

	position := 0
	for _, page := range g.pages {
		if page != nil {
			position = page.position + 1
		}
	}

	// ebitenutil.DebugPrint(screen, fmt.Sprintf("Page: %d, Y: %0.2f, Speed: %0.2f, TPS: %0.2f", g.position+1, g.y, g.speed, ebiten.CurrentFPS()))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Page: %d of %d (Loaded: %d), Speed: %0.2f, TPS: %0.2f", position, g.maxImages, len(g.images), g.speed, ebiten.CurrentTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}