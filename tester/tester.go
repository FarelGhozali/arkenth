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

// FuzzPage actively interacts with the page, trying edge cases in form fields.
func FuzzPage(page playwright.Page, rep *reporter.Report, originalURL string) {
	forms, err := page.Locator("form").All()
	if err != nil {
		log.Printf("Error finding forms on %s: %v", originalURL, err)
		return
	}

	if len(forms) == 0 {
		return
	}

	log.Printf("Found %d form(s) on %s, starting fuzzing...", len(forms), originalURL)

	for i := 0; i < len(forms); i++ {
		for _, payload := range fuzzPayloads {
			// Restore the original page state to properly access the form for each payload
			_, err := page.Goto(originalURL, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			})
			if err != nil {
				continue
			}

			// Re-query the form elements since DOM might have re-loaded
			formLocators, err := page.Locator("form").All()
			if err != nil || len(formLocators) <= i {
				continue
			}

			form := formLocators[i]
			inputs, err := form.Locator("input:not([type='hidden']):not([type='submit']), textarea").All()
			if err != nil || len(inputs) == 0 {
				continue
			}

			// Try to fill all applicable fields
			for _, input := range inputs {
				_ = input.Fill(payload)
			}

			// Submit the form
			submitBtns, err := form.Locator("button[type='submit'], input[type='submit']").All()
			if err == nil && len(submitBtns) > 0 {
				err = submitBtns[0].Click()
				if err != nil {
					log.Printf("Error clicking submit on form %d with payload %s: %v", i, payload, err)
				}
			} else {
				// Evaluate simple submit if no button
				_, _ = form.Evaluate("form => form.submit()", nil)
			}

			// Wait for potential navigation or error throwing
			page.WaitForTimeout(1000)
		}
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
