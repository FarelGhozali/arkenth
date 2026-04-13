package cmd

import (
	"log"
	"web-qa-automation/crawler"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:      "scan",
	Short:    "Start the automated QA scanning and fuzzing sequence",
	PreRunE:  RequireTarget,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting Enterprise Web QA Scan...")

		// Initialize spider/crawler with the global AppConfig loaded from rootCmd
		spider := crawler.NewSpider(AppConfig)
		err := spider.Run()
		if err != nil {
			log.Fatalf("Scan terminated with physical errors: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
