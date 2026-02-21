package cmd

import (
	"log"
	"web-qa-automation/crawler"
	"web-qa-automation/visual"

	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Run a visual crawl and compare current screenshots pixel-by-perfectly with /baseline",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting Visual Compare Crawl...")

		AppConfig.FastMode = true
		spider := crawler.NewSpider(AppConfig)

		spider.ProofDir = "./proofs/current"
		spider.SkipFuzzing = true

		err := spider.Run()
		if err != nil {
			log.Fatalf("Compare terminated with errors: %v", err)
		}

		// Proceed to Visual Diffing Engine
		log.Println("Navigating DOM structural diffing...")
		err = visual.GenerateRegressionReport("./proofs/baseline", "./proofs/current", "visual_regression_report.md")
		if err != nil {
			log.Fatalf("Failed generating visual report: %v", err)
		}

		log.Println("✅ Visual Regression complete. See visual_regression_report.md")
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)
}
