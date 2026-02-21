package visual

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateRegressionReport scans the baseline and current directories for matching images,
// runs the differ algorithm, and generates the markdown report.
func GenerateRegressionReport(baselineDir, currentDir, diffDir, reportFile string) error {
	os.MkdirAll(diffDir, 0755)

	f, err := os.Create(reportFile)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("# 👁️ Pixel-Perfect Visual Regression Report\n\n")

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

		diffPerc, err := CompareImages(baselineImg, currentImg, diffImg)
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
			f.WriteString("|----------|---------|-----------------------------|\n")

			// Resolve relative paths for md
			relBase, _ := filepath.Rel(".", baselineImg)
			relCurr, _ := filepath.Rel(".", currentImg)
			relDiff, _ := filepath.Rel(".", diffImg)

			f.WriteString(fmt.Sprintf("| ![%s](%s) | ![%s](%s) | ![%s](%s) |\n\n", "Baseline", relBase, "Current", relCurr, "Diff Image", relDiff))
		}
	}

	return nil
}
