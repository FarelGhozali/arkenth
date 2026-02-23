package a11y

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/playwright-community/playwright-go"
)

type AxeNode struct {
	Html   string   `json:"html"`
	Impact string   `json:"impact"`
	Target []string `json:"target"`
}

type AxeViolation struct {
	Description string    `json:"description"`
	Help        string    `json:"help"`
	HelpUrl     string    `json:"helpUrl"`
	Id          string    `json:"id"`
	Impact      string    `json:"impact"`
	Tags        []string  `json:"tags"`
	Nodes       []AxeNode `json:"nodes"`
}

type AxeReport struct {
	Violations []AxeViolation `json:"violations"`
}

// InitializeReport creates a fresh markdown report with a header
func InitializeReport(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("# ♿ Accessibility (WCAG) Audit Report\n\nThis report is automatically generated using `axe-core`. It highlights violations of the Web Content Accessibility Guidelines (WCAG).\n\n")
	return err
}

// RunAccessibilityScan fetches axe-core, injects it, and evaluates compliance.
func RunAccessibilityScan(page playwright.Page) ([]AxeViolation, error) {
	// 1. Inject axe-core library from CDN
	_, err := page.AddScriptTag(playwright.PageAddScriptTagOptions{
		URL: playwright.String("https://cdnjs.cloudflare.com/ajax/libs/axe-core/4.8.2/axe.min.js"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inject axe-core: %v", err)
	}

	// [Phase 2: Advanced] Dynamic State Mutation
	// Force-open hidden elements so the scanner can audit their contents
	_, err = page.Evaluate(`() => {
		// 1. Open all native <details> accordions
		document.querySelectorAll('details').forEach(d => d.setAttribute('open', 'true'));
		
		// 2. Attempt to trigger aria-expanded on common dropdown buttons
		document.querySelectorAll('[aria-expanded="false"]').forEach(el => {
			try { el.setAttribute('aria-expanded', 'true'); } catch(e) {}
		});
		
		// 3. Remove CSS display:none from common modal/dropdown classes (brute force visibility)
		const hiddenElements = document.querySelectorAll('.dropdown-menu, .modal, .offcanvas, [role="menu"]');
		hiddenElements.forEach(el => {
			try { el.style.display = 'block'; el.style.visibility = 'visible'; el.style.opacity = '1'; } catch(e) {}
		});
	}`)
	if err != nil {
		log.Printf("Warning: Failed to inject dynamic UI expander: %v", err)
	}

	// 2. Run the audit
	res, err := page.Evaluate(`async () => {
		return await axe.run();
	}`)
	if err != nil {
		return nil, fmt.Errorf("axe.run() failed: %v", err)
	}

	// 3. Convert map[string]interface{} to JSON then into Go Structs
	b, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal axe result: %v", err)
	}

	var axeRep AxeReport
	if err := json.Unmarshal(b, &axeRep); err != nil {
		return nil, fmt.Errorf("failed to unmarshal axe report: %v", err)
	}

	return axeRep.Violations, nil
}

// AppendViolations writes the findings for a specific URL into the report.
func AppendViolations(filename, url string, violations []AxeViolation) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("## 📄 URL: `%s`\n", url))
	if len(violations) == 0 {
		f.WriteString("✅ **Passed! No accessibility violations found.**\n\n")
		return nil
	}

	f.WriteString(fmt.Sprintf("❌ **Found %d violations.**\n\n", len(violations)))
	for _, v := range violations {
		f.WriteString(fmt.Sprintf("### Rule: [%s](%s) (%s Impact)\n", v.Id, v.HelpUrl, v.Impact))
		f.WriteString(fmt.Sprintf("**Description:** %s  \n", v.Description))
		f.WriteString(fmt.Sprintf("**Help:** %s  \n\n", v.Help))
		if len(v.Nodes) > 0 {
			f.WriteString("#### Failing Elements:\n")
			for _, n := range v.Nodes {
				target := "Unknown"
				if len(n.Target) > 0 {
					target = n.Target[0]
				}
				f.WriteString(fmt.Sprintf("- **Selector:** `%s`\n", target))
				f.WriteString(fmt.Sprintf("  - **HTML:** `%s`\n", n.Html))
			}
			f.WriteString("\n")
		}
	}
	f.WriteString("---\n\n")
	return nil
}
