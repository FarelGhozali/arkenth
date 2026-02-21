package cmd

import (
	"log"
	"web-qa-automation/crawler"

	"github.com/spf13/cobra"
)

var baselineCmd = &cobra.Command{
	Use:   "baseline",
	Short: "Take a pristine snapshot of the target URLs to act as visual baselines",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting Visual Baseline Crawl...")

		AppConfig.FastMode = true // Avoid heavy dynamic elements changing between snapshots
		spider := crawler.NewSpider(AppConfig)

		// Setup distinct folder for baselining
		spider.ProofDir = "./proofs/baseline"

		// Important: we don't want fuzzing during visual tests, just pure rendering
		spider.SkipFuzzing = true

		err := spider.Run()
		if err != nil {
			log.Fatalf("Baseline terminated with errors: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(baselineCmd)
}
