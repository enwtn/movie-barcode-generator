package imageprocess

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

func AveragePixels(row []Pixel) (p Pixel) {
	avg := Pixel{0, 0, 0}

	for _, pixel := range row {
		avg.add(pixel)
	}

	avg.R /= len(row)
	avg.G /= len(row)
	avg.B /= len(row)

	return avg
}

// CreateImage creates an image from a pixel array
func CreateImage(pixels [][]Pixel) {
	width := len(pixels[0])
	height := len(pixels)

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Set color for each pixel.
	for y := range pixels {
		for x := range pixels[0] {
			clr := color.RGBA{uint8(pixels[y][x].R), uint8(pixels[y][x].G), uint8(pixels[y][x].B), 255}
			img.Set(x, y, clr)
		}
	}

	// Encode as PNG.
	f, _ := os.Create("image.png")
	png.Encode(f, img)
}

// GetPixels produces a pixel array from the image file
func GetPixels(file io.Reader) ([][]Pixel, error) {
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		pixels = append(pixels, row)
	}

	return pixels, nil
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257)}
}

// Pixel struct example
type Pixel struct {
	R int
	G int
	B int
}

func addPixels(p [][]Pixel, p1 [][]Pixel) {
	for i := range p {
		for j := range p[0] {
			p[i][j].add(p1[i][j])
		}
	}
}

func (p *Pixel) add(p1 Pixel) {
	p.R += p1.R
	p.G += p1.G
	p.B += p1.B
}
