package crawler

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"web-qa-automation/models"

	"github.com/playwright-community/playwright-go"
)

var advancedPayloads = []string{
	// Boundary Values
	strings.Repeat("A", 20000), // Extreme length
	"-1",
	"0",
	"2147483647", // Max Int32
	// XSS
	"<script>alert(1)</script>",
	"\"><svg/onload=alert(1)>",
	// SQLi
	"' OR '1'='1",
	"admin' --",
	// Path Traversal
	"../../../etc/passwd",
	// Emojis / Unicode Parsing
	"👩‍👩‍👦‍👦🌟",
}

// Fuzzer interacts heavily with the DOM
type Fuzzer struct {
	Report *models.Report
}

func NewFuzzer(rep *models.Report) *Fuzzer {
	return &Fuzzer{Report: rep}
}

// FuzzAll applies payloads to all visible inputs
func (f *Fuzzer) FuzzAll(page playwright.Page, url string) {
	log.Printf("Starting advanced fuzzing on: %s", url)

	for _, payload := range advancedPayloads {
		// Attempt to inject payloads into every generic text/input field
		script := `(payload) => {
			let inputs = document.querySelectorAll("input:not([type='hidden']):not([type='submit']), textarea");
			let filled = 0;
			for (let i of inputs) {
				try {
					i.value = payload;
					i.dispatchEvent(new Event('input', { bubbles: true }));
					i.dispatchEvent(new Event('change', { bubbles: true }));
					filled++;
				} catch(e) {}
			}
			return filled;
		}`

		result, err := page.Evaluate(script, payload)
		if err != nil {
			continue
		}

		filledCount, ok := result.(int)
		if ok && filledCount > 0 {
			// Bruteforce click all buttons
			clickScript := `() => {
				let buttons = document.querySelectorAll("button, input[type='submit'], .btn");
				for (let b of buttons) {
					try { b.click(); } catch(e) {}
				}
			}`
			_, _ = page.Evaluate(clickScript)

			// Wait for potential crash or navigation
			page.WaitForTimeout(1500)
		}
	}
}

// CaptureCrash records a proof if an exception or crash is caught
func (f *Fuzzer) CaptureCrash(page playwright.Page, err error, proofDir string) {
	url := page.URL()
	filename := fmt.Sprintf("crash_%s.png", sanitizeFilename(url))
	proofPath := filepath.Join(proofDir, filename)

	_, screenshotErr := page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(proofPath),
		FullPage: playwright.Bool(true),
	})

	bug := models.Bug{
		Severity:    "CRITICAL",
		URL:         url,
		ActionTaken: "DOM Fuzzing or Page Load",
		Expected:    "Stable Page Render",
		Actual:      fmt.Sprintf("Caught Exception: %v", err),
	}

	if screenshotErr == nil {
		bug.ProofPath = proofPath
	}

	f.Report.CriticalBugs = append(f.Report.CriticalBugs, bug)
}

func sanitizeFilename(url string) string {
	replacer := strings.NewReplacer("https://", "", "http://", "", "/", "_", "?", "_", "&", "_", "=", "_", ":", "_")
	name := replacer.Replace(url)
	if len(name) > 50 { // Truncate to avoid OS file length limits
		name = name[:50]
	}
	return name
}
