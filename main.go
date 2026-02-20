package main

import (
	"log"

	"web-qa-automation/config"
	"web-qa-automation/crawler"
	"web-qa-automation/reporter"

	"github.com/playwright-community/playwright-go"
)

func main() {
	log.Println("Initializing Web QA Automation...")

	// 1. Parse CLI arguments
	cfg := config.ParseFlags()
	if cfg.TargetURL == "" {
		log.Fatalf("Please provide a --target-url to scan.")
	}

	// 2. Initialize Report
	rep := reporter.NewReport(cfg.TargetURL)

	// 3. Initialize Playwright Browser
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not install playwright drivers: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	// 4. Start Crawler & Modules
	log.Printf("Starting crawl for %s at depth %d", cfg.TargetURL, cfg.Depth)
	c := crawler.NewCrawler(cfg, rep, browser)
	err = c.Start()
	if err != nil {
		log.Fatalf("Error during crawling operations: %v", err)
	}

	// 5. Generate and Export Reports
	log.Println("Crawl finished. Generating reports...")

	err = rep.ExportJSON(cfg.OutputReport)
	if err != nil {
		log.Printf("Failed to export JSON report: %v", err)
	}

	err = rep.ExportMarkdown(cfg.OutputReport)
	if err != nil {
		log.Printf("Failed to export Markdown report: %v", err)
	}

	log.Printf("Execution successfully completed. Reports saved as %s.json and %s.md", cfg.OutputReport, cfg.OutputReport)
}
