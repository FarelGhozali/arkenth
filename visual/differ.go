package visual

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// CompareImages takes two image paths, compares them pixel-by-pixel, and generates a diff image.
// It returns the mismatch percentage and any error encountered.
func CompareImages(baselinePath, currentPath, diffOutputPath string) (float64, error) {
	// Open baseline image
	baseFile, err := os.Open(baselinePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open baseline: %v", err)
	}
	defer baseFile.Close()

	baseImg, err := png.Decode(baseFile)
	if err != nil {
		return 0, fmt.Errorf("failed to decode baseline PNG: %v", err)
	}

	// Open current image
	currFile, err := os.Open(currentPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open current: %v", err)
	}
	defer currFile.Close()

	currImg, err := png.Decode(currFile)
	if err != nil {
		return 0, fmt.Errorf("failed to decode current PNG: %v", err)
	}

	// Determine matching bounds, diff image needs to be the size of the larger image
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

	// Pixel by pixel comparison
	for y := 0; y < maxHeight; y++ {
		for x := 0; x < maxWidth; x++ {
			inB := x < bBounds.Max.X && y < bBounds.Max.Y
			inC := x < cBounds.Max.X && y < cBounds.Max.Y

			if inB && inC {
				// Both images have a pixel here
				rB, gB, bB, aB := baseImg.At(x, y).RGBA()
				rC, gC, bC, aC := currImg.At(x, y).RGBA()

				// If identical, faint background or distinct color based on requirement
				// A simple exact equality approach
				if rB == rC && gB == gC && bB == bC && aB == aC {
					// Draw gray-scale version of baseline to highlight diffs better
					gray := uint8((rB + gB + bB) / (3 * 257)) // 257 to scale from 16bit to 8bit
					gray = gray/2 + 128                       // wash out
					diffImg.Set(x, y, color.RGBA{R: gray, G: gray, B: gray, A: 255})
				} else {
					// Mismatch: paint it bright Red
					diffImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
					diffPixels++
				}
			} else {
				// Dimensions differ, missing pixel in one of the images
				diffImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				diffPixels++
			}
		}
	}

	// Save diff image
	diffFile, err := os.Create(diffOutputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create diff output file: %v", err)
	}
	defer diffFile.Close()

	if err := png.Encode(diffFile, diffImg); err != nil {
		return 0, fmt.Errorf("failed to encode diff image: %v", err)
	}

	diffPercentage := float64(diffPixels) / float64(totalPixels) * 100.0
	// Round to two decimal places
	return math.Round(diffPercentage*100) / 100, nil
}
