package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/FarelGhozali/arkenth/db"
	"github.com/FarelGhozali/arkenth/swagger"

	"github.com/spf13/cobra"
)

var (
	swaggerURL  string
	apiBaseURL  string
	projectFlag string
	concurrency int
)

var fuzzAPICmd = &cobra.Command{
	Use:   "fuzz-api",
	Short: "Perform smart fuzzing on a Swagger/OpenAPI endpoint",
	Long: `Automatically parse a Swagger/OpenAPI specification and generate
thousands of type-aware fuzz payloads for each endpoint.

Supports both local files and remote URLs for the spec:
  web-qa fuzz-api --swagger https://api.example.com/swagger.json --url https://api.example.com
  web-qa fuzz-api --swagger ./openapi.yaml --url http://localhost:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		if swaggerURL == "" {
			log.Fatal("❌ Please provide a Swagger/OpenAPI URL or file path using --swagger")
		}

		if apiBaseURL == "" {
			if AppConfig != nil && AppConfig.Target != "" {
				apiBaseURL = AppConfig.Target
			} else {
				log.Fatal("❌ Please provide a target API base URL using --url or --target")
			}
		}

		if projectFlag == "" {
			projectFlag = "default"
		}

		timestamp := time.Now().Format("20060102_150405")
		proofDir := filepath.Join("proofs", projectFlag, timestamp+"_fuzz_api")

		// Record the run in the database
		runID, _ := db.CreateRun(projectFlag, "fuzz-api", apiBaseURL, proofDir)

		log.Printf("🚀 Starting API Fuzzing for %s using spec %s", apiBaseURL, swaggerURL)

		// Phase 1: Parse the spec
		parser := swagger.NewParser()
		if err := parser.LoadSpec(swaggerURL); err != nil {
			log.Printf("❌ Failed to load Swagger spec: %v", err)
			db.UpdateRunStatus(runID, "failed")
			os.Exit(1)
		}

		endpoints, err := parser.ExtractEndpoints()
		if err != nil {
			log.Printf("❌ Failed to extract endpoints: %v", err)
			db.UpdateRunStatus(runID, "failed")
			os.Exit(1)
		}

		if len(endpoints) == 0 {
			log.Println("⚠️  No endpoints found in the spec. Nothing to fuzz.")
			db.UpdateRunStatus(runID, "completed")
			return
		}

		// Phase 2 & 3: Run the fuzzer
		fuzzer := swagger.NewFuzzer(apiBaseURL, endpoints, concurrency)
		fuzzer.Report.SpecURL = swaggerURL
		report := fuzzer.Run()

		// Phase 4: Generate report
		if err := swagger.GenerateReport(report, proofDir); err != nil {
			log.Printf("❌ Error generating report: %v", err)
			db.UpdateRunStatus(runID, "failed")
		} else {
			log.Printf("📄 Report generated at: %s", proofDir)
			db.UpdateRunStatus(runID, "completed")
		}

		// Print summary
		fmt.Println()
		fmt.Println("╔══════════════════════════════════════════════╗")
		fmt.Println("║     🔒 API Fuzzing Complete                  ║")
		fmt.Println("╠══════════════════════════════════════════════╣")
		fmt.Printf("║  📋 Endpoints:  %-5d                        ║\n", report.TotalEndpoints)
		fmt.Printf("║  📦 Requests:   %-5d                        ║\n", report.TotalRequests)
		fmt.Printf("║  🚨 Anomalies:  %-5d                        ║\n", report.TotalAnomalies)
		fmt.Printf("║  🔴 Critical:   %-5d                        ║\n", report.CriticalCount)
		fmt.Printf("║  🟠 High:       %-5d                        ║\n", report.HighCount)
		fmt.Printf("║  ⏱️  Duration:   %-6.1fs                      ║\n", report.DurationSeconds)
		fmt.Println("╚══════════════════════════════════════════════╝")
	},
}

func init() {
	fuzzAPICmd.Flags().StringVar(&swaggerURL, "swagger", "", "URL or path to swagger.json / openapi.yaml (required)")
	fuzzAPICmd.Flags().StringVar(&apiBaseURL, "url", "", "Base URL of the API to test (e.g., https://api.example.com)")
	fuzzAPICmd.Flags().StringVar(&projectFlag, "project", "default", "Project name for report organization")
	fuzzAPICmd.Flags().IntVar(&concurrency, "concurrency", 10, "Number of concurrent fuzz workers")
	rootCmd.AddCommand(fuzzAPICmd)
}
