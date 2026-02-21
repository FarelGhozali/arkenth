package cmd

import (
	"fmt"
	"os"

	"web-qa-automation/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AppConfig holds the finalized configuration for the execution
var AppConfig *config.AppConfig

var rootCmd = &cobra.Command{
	Use:   "web-qa",
	Short: "Enterprise Automated Web QA Testing CLI",
	Long: `A comprehensive web QA automation tool using Playwright-Go.
It supports deep network interception, fuzzing, mobile emulation, and proof generation.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Define global flags
	rootCmd.PersistentFlags().StringP("target", "t", "", "Target URL to scan (required)")
	rootCmd.PersistentFlags().IntP("depth", "d", 1, "Crawling depth (default is 1)")
	rootCmd.PersistentFlags().String("mobile-emulation", "", "Emulate a mobile device (e.g., 'iPhone 13')")
	rootCmd.PersistentFlags().String("auth-json", "", "Path to a Playwright storageState JSON for authenticated sessions")
	rootCmd.PersistentFlags().Bool("record-video", false, "Record test session videos to /proofs")
	rootCmd.PersistentFlags().Bool("fast-mode", false, "Block heavy media and tracking scripts to speed up scans")

	// Bind flags to Viper
	viper.BindPFlag("target", rootCmd.PersistentFlags().Lookup("target"))
	viper.BindPFlag("depth", rootCmd.PersistentFlags().Lookup("depth"))
	viper.BindPFlag("mobile-emulation", rootCmd.PersistentFlags().Lookup("mobile-emulation"))
	viper.BindPFlag("auth-json", rootCmd.PersistentFlags().Lookup("auth-json"))
	viper.BindPFlag("record-video", rootCmd.PersistentFlags().Lookup("record-video"))
	viper.BindPFlag("fast-mode", rootCmd.PersistentFlags().Lookup("fast-mode"))

	// Mark target as required
	rootCmd.MarkPersistentFlagRequired("target")
}

func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match

	AppConfig = config.LoadConfig()
	if AppConfig.Target != "" {
		fmt.Printf("Initialization Settings:\n- Target: %s\n- Depth: %d\n", AppConfig.Target, AppConfig.Depth)
	}
}
