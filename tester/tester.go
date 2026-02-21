package tester

import (
	"log"
	"strings"

	"web-qa-automation/reporter"

	"github.com/playwright-community/playwright-go"
)

// SetupPassiveMonitors hooks into page events to capture JS errors and network failures.
func SetupPassiveMonitors(page playwright.Page, rep *reporter.Report) {
	page.OnConsole(func(msg playwright.ConsoleMessage) {
		msgType := msg.Type()
		if msgType == "error" || msgType == "warning" {
			rep.AddJSFinding(reporter.JSFinding{
				URL:     page.URL(),
				Type:    msgType,
				Message: msg.Text(),
			})
		}
	})

	page.OnResponse(func(response playwright.Response) {
		status := response.Status()
		if status >= 400 {
			req := response.Request()
			rep.AddNetworkFinding(reporter.NetworkFinding{
				URL:    req.URL(),
				Method: req.Method(),
				Status: status,
				Error:  response.StatusText(),
			})
		}
	})

	page.OnRequestFailed(func(request playwright.Request) {
		errText := ""
		if request.Failure() != nil {
			errText = request.Failure().Error()
		}
		rep.AddNetworkFinding(reporter.NetworkFinding{
			URL:    request.URL(),
			Method: request.Method(),
			Status: 0,
			Error:  errText,
		})
	})
}

var fuzzPayloads = []string{
	"-1",
	"😊🚀",
	strings.Repeat("A", 10000),
	"<script>alert(1)</script>",
	"' OR 1=1 --",
}

// FuzzPage actively interacts with the page, trying edge cases across ALL inputs (even those without <form> tags).
func FuzzPage(page playwright.Page, rep *reporter.Report, originalURL string) {
	log.Printf("Starting aggressive fuzzing on ALL inputs for %s...", originalURL)

	for _, payload := range fuzzPayloads {
		// Restore the original page state to properly access elements for each payload
		_, err := page.Goto(originalURL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		if err != nil {
			continue
		}

		// Inject an aggressive JS script that fills ALL raw inputs in the DOM and clicks every button.
		// This bypasses the need for archaic <form> tags.
		fuzzScript := `(payload) => {
			let inputs = document.querySelectorAll("input:not([type='hidden']):not([type='submit']), textarea");
			let filled = 0;
			
			for (let i of inputs) {
				try {
					i.value = payload;
					i.dispatchEvent(new Event('input', { bubbles: true }));     // For React/Vue
					i.dispatchEvent(new Event('change', { bubbles: true }));
					filled++;
				} catch (e) {}
			}

			// If we filled something, let's just click EVERY button aggressively to trigger submitting
			if (filled > 0) {
				let buttons = document.querySelectorAll("button, input[type='submit'], input[type='button'], .btn");
				for (let btn of buttons) {
					try {
						btn.click();
					} catch (e) {}
				}
			}
			return filled;
		}`

		_, err = page.Evaluate(fuzzScript, payload)
		if err != nil {
			log.Printf("Error aggressively fuzzing inputs on %s with payload %s: %v", originalURL, payload, err)
		}

		// Wait briefly to allow potential SPA AJAX requests or navigation to fire before the next payload
		page.WaitForTimeout(2000)
	}
}

// MonitorPageError captures unhandled exceptions at the page level.
func MonitorPageError(page playwright.Page, rep *reporter.Report) {
	page.OnPageError(func(err error) {
		rep.AddVulnerability(reporter.VulnerabilityFinding{
			URL:      page.URL(),
			FormInfo: "Page Error / Crash",
			Payload:  "Generic Action Triggered",
			Error:    err.Error(),
		})
	})
}
