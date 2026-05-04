package visual

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FarelGhozali/arkenth/models"
)

// GenerateRegressionReport scans the baseline and current directories for matching images,
// runs the differ algorithm, and generates the markdown report.
func GenerateRegressionReport(baselineDir, currentDir, diffDir, reportFile string) error {
	return GenerateSmartRegressionReport(baselineDir, currentDir, diffDir, reportFile, nil, DefaultThreshold)
}

// GenerateSmartRegressionReport scans the baseline and current directories for matching images,
// runs the smart differ algorithm with masks and threshold, and generates the markdown report.
func GenerateSmartRegressionReport(baselineDir, currentDir, diffDir, reportFile string, masks []models.VisualMask, threshold float64) error {
	os.MkdirAll(diffDir, 0755)

	cleanReportFile := filepath.Clean(reportFile)
	f, err := os.Create(cleanReportFile)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("# 🧠 Smart Visual Regression Report\n\n")
	f.WriteString(fmt.Sprintf("> **Threshold:** %.1f (Euclidean color distance)  \n", threshold))
	f.WriteString(fmt.Sprintf("> **Active Masks:** %d ignore region(s)  \n\n", len(masks)))

	files, err := os.ReadDir(baselineDir)
	if err != nil {
		return fmt.Errorf("baseline directory missing: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".png") {
			continue
		}

		baselineImg := filepath.Join(baselineDir, file.Name())
		currentImg := filepath.Join(currentDir, file.Name())
		diffImg := filepath.Join(diffDir, "diff_"+file.Name())

		if _, err := os.Stat(currentImg); os.IsNotExist(err) {
			f.WriteString(fmt.Sprintf("## ❌ Missing Page: `%s`\n", file.Name()))
			f.WriteString("> This page existed in baseline but was not found during the current crawl.\n\n")
			continue
		}

		diffPerc, err := SmartCompareImages(baselineImg, currentImg, diffImg, masks, threshold)
		if err != nil {
			f.WriteString(fmt.Sprintf("## ⚠️ Error comparing `%s`: %v\n\n", file.Name(), err))
			continue
		}

		status := "✅ MATCH (0% Diff)"
		// Any diff larger than 0.00% is flagged
		if diffPerc > 0 {
			status = fmt.Sprintf("❌ MISMATCH (%.2f%% Diff)", diffPerc)
		}

		f.WriteString(fmt.Sprintf("## File: `%s` -> %s\n", file.Name(), status))

		if diffPerc > 0 {
			f.WriteString("### Visual Evidence\n")

			// Simple layout to show baseline -> current -> diff horizontally if possible
			// Markdown tables work best for image side-by-sides
			f.WriteString("| Baseline | Current | Visual Diff (Red = Changed) |\n")
			f.WriteString("|----------|---------|-----------------------------|")
			f.WriteString("\n")

			// Resolve relative paths for md
			relBase, _ := filepath.Rel(".", baselineImg)
			relCurr, _ := filepath.Rel(".", currentImg)
			relDiff, _ := filepath.Rel(".", diffImg)

			f.WriteString(fmt.Sprintf("| ![%s](%s) | ![%s](%s) | ![%s](%s) |\n\n", "Baseline", relBase, "Current", relCurr, "Diff Image", relDiff))
		}
	}

	return nil
}
