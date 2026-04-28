package cmd

import (
	"github.com/FarelGhozali/arkenth/a11y"
	"github.com/FarelGhozali/arkenth/crawler"
	"log"

	"github.com/spf13/cobra"
)

var a11yCmd = &cobra.Command{
	Use:     "a11y",
	Short:   "Run an accessibility (WCAG) audit on the target URLs",
	PreRunE: RequireTarget,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting Accessibility Audit...")

		// Initialize the report file
		reportFile := "accessibility_audit_report.md"
		err := a11y.InitializeReport(reportFile)
		if err != nil {
			log.Fatalf("Failed to initialize a11y report: %v", err)
		}

		AppConfig.FastMode = true
		spider := crawler.NewSpider(AppConfig)

		// Configure spider for a11y scan
		spider.SkipFuzzing = true
		spider.RunA11y = true

		err = spider.Run()
		if err != nil {
			log.Fatalf("A11y scan terminated with errors: %v", err)
		}

		log.Println("✅ Accessibility Audit complete. See accessibility_audit_report.md")
	},
}

func init() {
	rootCmd.AddCommand(a11yCmd)
}
