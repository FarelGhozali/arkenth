package visual

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/FarelGhozali/web-qa-automation/models"
)

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func comparePixel(baseImg, currImg image.Image, x, y int, bBounds, cBounds image.Rectangle) (color.Color, bool) {
	inB := x < bBounds.Max.X && y < bBounds.Max.Y
	inC := x < cBounds.Max.X && y < cBounds.Max.Y

	if inB && inC {
		rB, gB, bB, aB := baseImg.At(x, y).RGBA()
		rC, gC, bC, aC := currImg.At(x, y).RGBA()

		if rB == rC && gB == gC && bB == bC && aB == aC {
			gray := uint8((rB + gB + bB) / (3 * 257))
			gray = gray/2 + 128
			return color.RGBA{R: gray, G: gray, B: gray, A: 255}, false
		}
	}
	return color.RGBA{R: 255, G: 0, B: 0, A: 255}, true
}

// CompareImages takes two image paths, compares them pixel-by-pixel, and generates a diff image.
// It returns the mismatch percentage and any error encountered.
func CompareImages(baselinePath, currentPath, diffOutputPath string) (float64, error) {
	baseImg, err := loadImage(baselinePath)
	if err != nil {
		return 0, fmt.Errorf("failed to load baseline: %v", err)
	}

	currImg, err := loadImage(currentPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load current: %v", err)
	}

	bBounds := baseImg.Bounds()
	cBounds := currImg.Bounds()

	maxWidth := bBounds.Max.X
	if cBounds.Max.X > maxWidth {
		maxWidth = cBounds.Max.X
	}

	maxHeight := bBounds.Max.Y
	if cBounds.Max.Y > maxHeight {
		maxHeight = cBounds.Max.Y
	}

	diffImg := image.NewRGBA(image.Rect(0, 0, maxWidth, maxHeight))

	diffPixels := 0
	totalPixels := maxWidth * maxHeight

	for y := 0; y < maxHeight; y++ {
		for x := 0; x < maxWidth; x++ {
			c, isDiff := comparePixel(baseImg, currImg, x, y, bBounds, cBounds)
			diffImg.Set(x, y, c)
			if isDiff {
				diffPixels++
			}
		}
	}

	diffFile, err := os.Create(diffOutputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create diff output file: %v", err)
	}
	defer diffFile.Close()

	if err := png.Encode(diffFile, diffImg); err != nil {
		return 0, fmt.Errorf("failed to encode diff image: %v", err)
	}

	diffPercentage := float64(diffPixels) / float64(totalPixels) * 100.0
	return math.Round(diffPercentage*100) / 100, nil
}

// colorDistance computes the Euclidean distance between two RGB colors.
// This is the core of the "Dynamic Threshold" feature: small rendering
// differences (anti-aliasing, GPU variance) produce distances < ~10,
// while genuine visual changes produce distances > 30.
func colorDistance(r1, g1, b1, r2, g2, b2 uint8) float64 {
	dr := float64(r1) - float64(r2)
	dg := float64(g1) - float64(g2)
	db := float64(b1) - float64(b2)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// toRGBA converts any image.Image to an *image.RGBA so that
// we can mutate individual pixels (needed for mask painting).
func toRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}

// applyMasks fills each mask region with solid black (0,0,0) pixels.
// This effectively "erases" dynamic areas from comparison.
func applyMasks(img *image.RGBA, masks []models.VisualMask) {
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	bounds := img.Bounds()

	for _, m := range masks {
		for y := m.Y; y < m.Y+m.Height && y < bounds.Max.Y; y++ {
			for x := m.X; x < m.X+m.Width && x < bounds.Max.X; x++ {
				if x >= bounds.Min.X && y >= bounds.Min.Y {
					img.Set(x, y, black)
				}
			}
		}
	}
}
