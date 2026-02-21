package crawler

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"web-qa-automation/config"
	"web-qa-automation/interceptor"
	"web-qa-automation/models"
	"web-qa-automation/reporter"

	"github.com/playwright-community/playwright-go"
)

type Spider struct {
	Config  *config.AppConfig
	Report  *models.Report
	Visited map[string]bool
	Fuzzer  *Fuzzer
}

func NewSpider(cfg *config.AppConfig) *Spider {
	mobile := false
	if cfg.MobileEmulation != "" {
		mobile = true
	}
	rep := models.NewReport(cfg.Target, mobile)
	return &Spider{
		Config:  cfg,
		Report:  rep,
		Visited: make(map[string]bool),
		Fuzzer:  NewFuzzer(rep),
	}
}

// Run boots Playwright, handles context setup (Auth/Mobile), triggers crawl, and kicks off Reporter
func (s *Spider) Run() error {
	proofDir := "./proofs"
	f, _ := os.Stat(proofDir)
	if f == nil {
		os.Mkdir(proofDir, 0755)
	}

	err := playwright.Install()
	if err != nil {
		log.Printf("Playwright installation notice: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return err
	}
	defer browser.Close()

	// Handle Context Options (Video, Emulation, Auth State)
	ctxOptions := playwright.BrowserNewContextOptions{
		IgnoreHttpsErrors: playwright.Bool(true),
	}

	if s.Config.RecordVideo {
		ctxOptions.RecordVideo = &playwright.RecordVideo{
			Dir: proofDir,
		}
	}

	if s.Config.AuthJSON != "" {
		ctxOptions.StorageState = playwright.String(s.Config.AuthJSON)
		log.Printf("Loaded Authentication State from %s", s.Config.AuthJSON)
	}

	if s.Config.MobileEmulation != "" {
		if device, ok := pw.Devices[s.Config.MobileEmulation]; ok {
			ctxOptions.Viewport = device.Viewport
			ctxOptions.UserAgent = playwright.String(device.UserAgent)
			ctxOptions.DeviceScaleFactor = playwright.Float(device.DeviceScaleFactor)
			ctxOptions.IsMobile = playwright.Bool(device.IsMobile)
			ctxOptions.HasTouch = playwright.Bool(device.HasTouch)
			log.Printf("Emulating Device: %s", s.Config.MobileEmulation)
		} else {
			log.Printf("Warning: Device profile '%s' not found. Reverting to Desktop.", s.Config.MobileEmulation)
		}
	}

	// Begin Recursive Crawl Core
	s.crawl(s.Config.Target, 0, browser, ctxOptions, proofDir)

	// Clean up videos if flag wasn't sent, or process outputs
	if !s.Config.RecordVideo {
		// Video directories get created inherently sometimes if not explicitly handled per page.
		os.RemoveAll("./proofs")
		os.Mkdir("./proofs", 0755) // Recreate fresh for purely screenshots
	}

	log.Println("Crawl Complete. Generating Reports...")

	rep := reporter.NewReporter(s.Report, proofDir)
	if err := rep.GenerateOutputs(); err != nil {
		log.Printf("Error generating reports: %v", err)
	} else {
		log.Println("✅ Run completed successfully. Check 'qa_audit_report.md' and the 'proofs' folder.")
	}

	return nil
}

func (s *Spider) crawl(targetURL string, currentDepth int, browser playwright.Browser, ctxOpts playwright.BrowserNewContextOptions, proofDir string) {
	if currentDepth > s.Config.Depth {
		return
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return
	}
	parsedURL.Fragment = "" // Strip anchors
	cleanURL := parsedURL.String()

	if s.Visited[cleanURL] {
		return
	}
	s.Visited[cleanURL] = true
	s.Report.ScannedURLs = append(s.Report.ScannedURLs, cleanURL)
	log.Printf("🕷️ Crawling [%d/%d]: %s", currentDepth, s.Config.Depth, cleanURL)

	// Spin isolated context
	context, err := browser.NewContext(ctxOpts)
	if err != nil {
		log.Printf("Failed to create context: %v", err)
		return
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return
	}

	// 1. Setup Interceptor
	intercep := interceptor.NewInterceptor(s.Config, s.Report)
	intercep.AttachToPage(page)

	// Check page crashes aggressively
	page.OnPageError(func(err error) {
		s.Fuzzer.CaptureCrash(page, err, proofDir)
	})

	// 2. Navigation
	_, err = page.Goto(cleanURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})

	if err != nil {
		log.Printf("Navigation failed on %s: %v", cleanURL, err)
		return
	}

	// 3. Take Proof Screenshot of the unique URL view
	filename := fmt.Sprintf("view_%s.png", sanitizeFilename(cleanURL))
	page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(filepath.Join(proofDir, filename)),
		FullPage: playwright.Bool(true),
	})

	// 4. Advanced Fuzzing Phase
	s.Fuzzer.FuzzAll(page, cleanURL)

	// 5. Gather subsequent internal links for next Depth Layer
	if currentDepth < s.Config.Depth {
		// Re-navigate cleanly to grab links uncontaminated by fuzz states
		cleanPage, _ := context.NewPage()
		cleanPage.Goto(cleanURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})

		links, _ := extractInternalLinks(cleanPage, cleanURL)
		cleanPage.Close()

		for _, link := range links {
			s.crawl(link, currentDepth+1, browser, ctxOpts, proofDir)
		}
	}
}

func extractInternalLinks(page playwright.Page, baseURL string) ([]string, error) {
	hrefs, err := page.Locator("a[href]").EvaluateAll(`elements => elements.map(e => e.href)`)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	var internalLinks []string
	visitedLocal := make(map[string]bool)

	if hrefsSlice, ok := hrefs.([]interface{}); ok {
		for _, h := range hrefsSlice {
			if strUrl, ok := h.(string); ok {
				parsed, err := url.Parse(strUrl)
				if err != nil {
					continue
				}

				if parsed.Host == base.Host || parsed.Host == "" {
					parsed.Fragment = ""
					absURL := base.ResolveReference(parsed).String()

					if parsed.Scheme == "javascript" || parsed.Scheme == "mailto" {
						continue
					}

					if !visitedLocal[absURL] {
						visitedLocal[absURL] = true
						internalLinks = append(internalLinks, absURL)
					}
				}
			}
		}
	}

	return internalLinks, nil
}
