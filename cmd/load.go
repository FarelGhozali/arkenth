package cmd

import (
	"log"
	"time"
	"web-qa-automation/loadtester"

	"github.com/spf13/cobra"
)

var (
	users    int
	duration string
	method   string
	bodyJson string
)

var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "Run a highly concurrent HTTP load test against the target",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Starting Load Test on %s with %d users...", AppConfig.Target, users)

		parsedDuration, err := time.ParseDuration(duration)
		if err != nil {
			log.Fatalf("Invalid duration format. Use formats like '10s', '1m'. Error: %v", err)
		}

		res := loadtester.RunLoadTest(AppConfig.Target, users, parsedDuration, method, bodyJson)

		log.Println("Load Test Completed. Generating report...")
		err = loadtester.GenerateLoadReport("load_test_report.md", AppConfig.Target, users, res)
		if err != nil {
			log.Fatalf("Failed to generate report: %v", err)
		}

		log.Println("✅ Performance & Load Testing complete. See load_test_report.md")
	},
}

func init() {
	loadCmd.Flags().IntVar(&users, "users", 50, "Number of concurrent virtual users")
	loadCmd.Flags().StringVar(&duration, "duration", "10s", "Duration of the load test (e.g. 10s, 1m)")
	loadCmd.Flags().StringVar(&method, "method", "GET", "HTTP Method to use (GET, POST, PUT, DELETE)")
	loadCmd.Flags().StringVar(&bodyJson, "body-json", "", "JSON string to send as request body. Use {{RANDOM}} for cache-busting mutations.")
	rootCmd.AddCommand(loadCmd)
}
