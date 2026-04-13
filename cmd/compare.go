package cmd

import (
	"fmt"
	"log"
	"time"
	"web-qa-automation/crawler"
	"web-qa-automation/visual"

	"github.com/spf13/cobra"
)

var baselineDate string

var compareCmd = &cobra.Command{
	Use:      "compare",
	Short:    "Run a visual crawl and compare current screenshots pixel-by-perfectly with /baseline",
	PreRunE:  RequireTarget,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting Visual Compare Crawl...")

		AppConfig.FastMode = true
		spider := crawler.NewSpider(AppConfig)

		dateStr := time.Now().Format("02-01-2006")
		spider.ProofDir = fmt.Sprintf("./proofs/%s/current", dateStr)
		spider.SkipFuzzing = true

		err := spider.Run()
		if err != nil {
			log.Fatalf("Compare terminated with errors: %v", err)
		}

		// Proceed to Visual Diffing Engine
		log.Println("Navigating DOM structural diffing...")
		bDate := baselineDate
		if bDate == "" {
			bDate = dateStr
		}
		baselineDir := fmt.Sprintf("./proofs/%s/baseline", bDate)
		diffDir := fmt.Sprintf("./proofs/%s/diff", dateStr)
		err = visual.GenerateRegressionReport(baselineDir, spider.ProofDir, diffDir, "visual_regression_report.md")
		if err != nil {
			log.Fatalf("Failed generating visual report: %v", err)
		}

		log.Println("✅ Visual Regression complete. See visual_regression_report.md")
	},
}

func init() {
	compareCmd.Flags().StringVar(&baselineDate, "baseline-date", "", "Date of the baseline to compare against (DD-MM-YYYY). Defaults to today.")
	rootCmd.AddCommand(compareCmd)
}
