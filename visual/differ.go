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

// DefaultThreshold is the default Euclidean color distance tolerance.
// Differences below this value are considered identical, eliminating
// false positives caused by sub-pixel rendering or anti-aliasing.
const DefaultThreshold = 12.0

// CompareImages takes two image paths, compares them pixel-by-pixel, and generates a diff image.
// It returns the mismatch percentage and any error encountered.
func CompareImages(baselinePath, currentPath, diffOutputPath string) (float64, error) {
	return SmartCompareImages(baselinePath, currentPath, diffOutputPath, nil, DefaultThreshold)
}

// SmartCompareImages performs a tolerance-aware, mask-aware visual comparison.
//   - masks: slice of VisualMask regions to ignore (filled with solid color before comparison).
//   - threshold: Euclidean color distance below which two pixels are considered identical.
//     Pass 0 for exact pixel matching (legacy behaviour).
func SmartCompareImages(baselinePath, currentPath, diffOutputPath string, masks []models.VisualMask, threshold float64) (float64, error) {
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

	// Apply masks: draw solid black rectangles on both images in the masked areas
	baseRGBA := toRGBA(baseImg)
	currRGBA := toRGBA(currImg)
	applyMasks(baseRGBA, masks)
	applyMasks(currRGBA, masks)

	// Determine matching bounds, diff image needs to be the size of the larger image
	bBounds := baseRGBA.Bounds()
	cBounds := currRGBA.Bounds()

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

	// Pixel by pixel comparison with Euclidean color distance
	for y := 0; y < maxHeight; y++ {
		for x := 0; x < maxWidth; x++ {
			inB := x < bBounds.Max.X && y < bBounds.Max.Y
			inC := x < cBounds.Max.X && y < cBounds.Max.Y

			if inB && inC {
				// Both images have a pixel here
				rB, gB, bB, _ := baseRGBA.At(x, y).RGBA()
				rC, gC, bC, _ := currRGBA.At(x, y).RGBA()

				// Scale from 16-bit to 8-bit for Euclidean distance
				dist := colorDistance(
					uint8(rB>>8), uint8(gB>>8), uint8(bB>>8),
					uint8(rC>>8), uint8(gC>>8), uint8(bC>>8),
				)

				if dist <= threshold {
					// Within tolerance — draw faint gray background
					gray := uint8((rB + gB + bB) / (3 * 257))
					gray = gray/2 + 128
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
