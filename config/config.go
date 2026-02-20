package config

import "flag"

type Config struct {
	TargetURL    string
	Depth        int
	OutputReport string
}

func ParseFlags() *Config {
	targetURL := flag.String("target-url", "", "The target URL to scan")
	depth := flag.Int("depth", 1, "The crawling depth")
	outputReport := flag.String("output-report", "report", "Prefix for the output report files (e.g., 'report' will generate 'report.json' and 'report.md')")
	flag.Parse()

	return &Config{
		TargetURL:    *targetURL,
		Depth:        *depth,
		OutputReport: *outputReport,
	}
}
