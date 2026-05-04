package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FarelGhozali/arkenth/a11y"
	"github.com/FarelGhozali/arkenth/config"
	"github.com/FarelGhozali/arkenth/crawler"
	"github.com/FarelGhozali/arkenth/db"
	"github.com/FarelGhozali/arkenth/loadtester"
	"github.com/FarelGhozali/arkenth/swagger"
	"github.com/FarelGhozali/arkenth/visual"
)

// --- Execution Helpers (run in goroutines) ---

func executeScan(cfg *config.AppConfig, projectName string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_scan", projectName, timestamp)

	runID, _ := db.CreateRun(projectName, "scan", cfg.Target, proofDir)
	log.Printf("[UI] Starting Scan on %s (Project: %s) ...", cfg.Target, projectName)

	spider := crawler.NewSpider(cfg)
	spider.ProofDir = proofDir

	if err := spider.Run(); err != nil {
		log.Printf("[UI] Scan error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Scan completed successfully.")
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeBaseline(cfg *config.AppConfig, projectName string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_baseline", projectName, timestamp)

	runID, _ := db.CreateRun(projectName, "baseline", cfg.Target, proofDir)
	log.Printf("[UI] Starting Baseline on %s (Project: %s) ...", cfg.Target, projectName)

	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	spider.ProofDir = proofDir
	spider.SkipFuzzing = true
	if err := spider.Run(); err != nil {
		log.Printf("[UI] Baseline error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Baseline completed successfully.")
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeCompare(cfg *config.AppConfig, projectName string, baselinePath string, threshold float64) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_compare", projectName, timestamp)

	runID, _ := db.CreateRun(projectName, "compare", cfg.Target, proofDir)
	log.Printf("[UI] Starting Smart Compare on %s (Project: %s, Threshold: %.1f) ...", cfg.Target, projectName, threshold)

	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	spider.ProofDir = proofDir
	spider.SkipFuzzing = true

	if err := spider.Run(); err != nil {
		log.Printf("[UI] Compare crawl error: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	baselineDir := baselinePath
	if !strings.Contains(baselineDir, "/") && !strings.Contains(baselineDir, "\\") {
		baselineDir = filepath.Join("proofs", projectName, baselinePath)
	}

	// Set threshold default if not specified
	if threshold <= 0 {
		threshold = visual.DefaultThreshold
	}

	// Fetch masks from database for the target URL
	masks, err := db.GetVisualMasks(cfg.Target)
	if err != nil {
		log.Printf("[UI] Warning: Could not load visual masks: %v", err)
		masks = nil
	}

	diffDir := filepath.Join(spider.ProofDir, "diff")
	reportPath := filepath.Join(spider.ProofDir, "visual_regression_report.md")
	if err := visual.GenerateSmartRegressionReport(baselineDir, spider.ProofDir, diffDir, reportPath, masks, threshold); err != nil {
		log.Printf("[UI] Visual diff error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Printf("✅ Smart Visual Regression complete. Threshold=%.1f, Masks=%d", threshold, len(masks))
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeA11y(cfg *config.AppConfig, projectName string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_a11y", projectName, timestamp)

	runID, _ := db.CreateRun(projectName, "a11y", cfg.Target, proofDir)
	log.Printf("[UI] Starting A11y Audit on %s (Project: %s) ...", cfg.Target, projectName)

	os.MkdirAll(proofDir, 0755)
	reportFile := filepath.Join(proofDir, "accessibility_audit_report.md")
	if err := a11y.InitializeReport(reportFile); err != nil {
		log.Printf("[UI] A11y init error: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	cfg.FastMode = true
	spider := crawler.NewSpider(cfg)
	spider.ProofDir = proofDir
	spider.SkipFuzzing = true
	spider.RunA11y = true
	if err := spider.Run(); err != nil {
		log.Printf("[UI] A11y error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Accessibility Audit completed.")
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeLoad(cfg *config.AppConfig, projectName string, users int, duration string, method string, bodyJSON string) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_load", projectName, timestamp)

	runID, _ := db.CreateRun(projectName, "load", cfg.Target, proofDir)
	log.Printf("[UI] Starting Load Test on %s (Project: %s) with %d users...", cfg.Target, projectName, users)

	if users <= 0 {
		users = 50
	}
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		log.Printf("[UI] Invalid duration: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	os.MkdirAll(proofDir, 0755)
	res := loadtester.RunLoadTest(cfg.Target, users, parsedDuration, method, bodyJSON)

	reportPath := filepath.Join(proofDir, "load_test_report.md")
	if err := loadtester.GenerateLoadReport(reportPath, cfg.Target, users, res); err != nil {
		log.Printf("[UI] Load report error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Println("✅ Load Test completed.")
		db.UpdateRunStatus(runID, "completed")
	}
}

func executeFuzzAPI(projectName string, targetBaseURL string, swaggerURL string, concurrencyLevel int) {
	timestamp := time.Now().Format("20060102_150405")
	proofDir := fmt.Sprintf("./proofs/%s/%s_fuzz_api", projectName, timestamp)

	runID, _ := db.CreateRun(projectName, "fuzz-api", targetBaseURL, proofDir)
	log.Printf("[UI] Starting API Fuzzing on %s using spec %s (Project: %s) ...", targetBaseURL, swaggerURL, projectName)

	parser := swagger.NewParser()
	if err := parser.LoadSpec(swaggerURL); err != nil {
		log.Printf("[UI] Failed to load Swagger spec: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	endpoints, err := parser.ExtractEndpoints()
	if err != nil {
		log.Printf("[UI] Failed to extract endpoints: %v", err)
		db.UpdateRunStatus(runID, "failed")
		return
	}

	if concurrencyLevel <= 0 {
		concurrencyLevel = 10
	}

	fuzzer := swagger.NewFuzzer(targetBaseURL, endpoints, concurrencyLevel)
	fuzzer.Report.SpecURL = swaggerURL
	report := fuzzer.Run()

	if err := swagger.GenerateReport(report, proofDir); err != nil {
		log.Printf("[UI] API Fuzz report error: %v", err)
		db.UpdateRunStatus(runID, "failed")
	} else {
		log.Printf("✅ API Fuzzing completed. %d anomalies found.", report.TotalAnomalies)
		db.UpdateRunStatus(runID, "completed")
	}
}
