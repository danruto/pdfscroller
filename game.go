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
	screenHeight = 600
)

// NewImageFromReader creates a new image from a data stream.
// The name parameter is required to uniquely identify this image (for caching etc).
// If the image in this io.Reader is an SVG, the name should end ".svg".
// Images returned from this method will scale to fit the canvas object.
// The method for scaling can be set using the Fill field.
//
// Since: 2.0
func NewImageFromReader(read io.Reader, name string) *image.Image {
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

	// TODO: We can use name for a side panel later if we want
	return &ret
}

type Game struct {
	images   []*ebiten.Image
	position int
	y        float64
	speed    float64
}

func NewGame() (*Game, error) {
	f, err := os.Open("chapter46.pdf")
	if err != nil {
		return nil, err
	}

	c, err := pdfapi.PageCount(f, nil)
	if err != nil {
		return nil, err
	}
	log.Println(fmt.Sprintf("PageCount: %d", c))

	images, err := pdfapi.ExtractImagesRaw(f, nil, nil)
	if err != nil {
		return nil, err
	}
	log.Println(fmt.Sprintf("ExtractImagesRaw: %d", len(images)))

	var gameImages []*ebiten.Image
	for _, image := range images {
		img := NewImageFromReader(image.Reader, image.Name)
		log.Println("Added an image into the box")

		gameImages = append(gameImages, ebiten.NewImageFromImage(*img))
	}

	g := &Game{
		images:   gameImages,
		position: 0,
		y:        0,
		speed:    1.0,
	}

	return g, nil
}

func (g *Game) Update() error {

	// If our y is approximately at the border of the next image, we render 2 images
	absY := math.Abs(g.y)
	maxY := float64(g.images[g.position].Bounds().Size().Y)
	halfScreenY := float64(screenHeight / 2)
	bounds := maxY - screenHeight
	if absY >= bounds {
		if g.position < len(g.images)-1 {
			g.y = halfScreenY
			g.position += 1
			log.Println("Updating position", absY, maxY, halfScreenY, bounds)
		}
	} else {
		g.y -= g.speed
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		g.speed -= 0.2
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		g.speed += 0.2
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Only render the image window based on the y and position
	// TODO: The rolling 2 images
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, g.y)
	screen.DrawImage(g.images[g.position], op)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("Page: %d, Y: %0.2f, Speed: %0.2f, TPS: %0.2f", g.position+1, g.y, g.speed, ebiten.CurrentFPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}