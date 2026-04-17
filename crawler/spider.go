package crawler

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/FarelGhozali/web-qa-automation/a11y"
	"github.com/FarelGhozali/web-qa-automation/config"
	"github.com/FarelGhozali/web-qa-automation/interceptor"
	"github.com/FarelGhozali/web-qa-automation/models"
	"github.com/FarelGhozali/web-qa-automation/reporter"

	"github.com/playwright-community/playwright-go"
)

type Spider struct {
	Config      *config.AppConfig
	Report      *models.Report
	Visited     map[string]bool
	Fuzzer      *Fuzzer
	ProofDir    string // Control where images go (default 'proofs', visual: 'proofs/baseline' or 'proofs/current')
	SkipFuzzing bool   // Control if active fuzzing happens (visual baseline only wants layouts)
	RunA11y     bool   // Control if Accessibility checks should run
}

func NewSpider(cfg *config.AppConfig) *Spider {
	mobile := false
	if cfg.MobileEmulation != "" {
		mobile = true
	}
	rep := models.NewReport(cfg.Target, mobile)
	return &Spider{
		Config:      cfg,
		Report:      rep,
		Visited:     make(map[string]bool),
		Fuzzer:      NewFuzzer(rep),
		ProofDir:    fmt.Sprintf("./proofs/%s/scan", time.Now().Format("02-01-2006")),
		SkipFuzzing: false,
		RunA11y:     false,
	}
}

// Run boots Playwright, handles context setup (Auth/Mobile), triggers crawl, and kicks off Reporter
func (s *Spider) Run() error {
	f, _ := os.Stat(s.ProofDir)
	if f == nil {
		os.MkdirAll(s.ProofDir, 0755)
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
			Dir: s.ProofDir,
		}
	}

	if s.Config.AuthJSON != "" {
		ctxOptions.StorageStatePath = playwright.String(s.Config.AuthJSON)
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
	s.crawl(s.Config.Target, 0, browser, ctxOptions, s.ProofDir)

	// Clean up videos if flag wasn't sent, or process outputs
	if !s.Config.RecordVideo {
		// Just ensure director is ok, won't aggressively delete anymore, just rely on new images replacing
	}

	log.Println("Crawl Complete. Generating Reports...")

	rep := reporter.NewReporter(s.Report, s.ProofDir)
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

	// 2. Navigation - Use Networkidle for SPAs (React/Next.js) to finish fetching APIs and hiding loading screens
	_, err = page.Goto(cleanURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})

	if err != nil {
		log.Printf("Navigation failed on %s: %v", cleanURL, err)
		return
	}

	// Hard pause to allow CSS animations / Modals to settle
	time.Sleep(2 * time.Second)

	// 3. Take Proof Screenshot of the unique URL view
	filename := fmt.Sprintf("view_%s.png", sanitizeFilename(cleanURL))
	page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(filepath.Join(proofDir, filename)),
		FullPage: playwright.Bool(true),
	})

	// 4. Advanced Fuzzing Phase
	if !s.SkipFuzzing {
		s.Fuzzer.FuzzAll(page, cleanURL)
		s.Fuzzer.TestTokenTampering(page, cleanURL)
	} else if !s.RunA11y {
		log.Printf("📸 Visual mode: Snapping [%s] cleanly without DOM fuzzing.", cleanURL)
	}

	// 5. Accessibility Scan
	if s.RunA11y {
		log.Printf("♿ Running a11y rules against %s", cleanURL)
		violations, err := a11y.RunAccessibilityScan(page)
		if err != nil {
			log.Printf("A11y Error on %s: %v", cleanURL, err)
		} else {
			a11y.AppendViolations("accessibility_audit_report.md", cleanURL, violations)
		}
	}

	// 6. Gather subsequent internal links for next Depth Layer
	if currentDepth < s.Config.Depth {
		// Re-navigate cleanly to grab links uncontaminated by fuzz states
		cleanPage, _ := context.NewPage()
		cleanPage.Goto(cleanURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})

		links, _ := extractInternalLinks(cleanPage, cleanURL)

		// [Phase 2: Advanced Feature] - Smart JS Hunt
		hiddenAPIs, _ := scrapeJSForHiddenEndpoints(cleanPage, cleanURL)
		if len(hiddenAPIs) > 0 {
			log.Printf("🕵️  Smart Spider: Discovered %d hidden endpoints from Javascript on %s", len(hiddenAPIs), cleanURL)
		}

		links = append(links, hiddenAPIs...)
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

// scrapeJSForHiddenEndpoints downloads JS files and uses Regex to find hardcoded API routes
func scrapeJSForHiddenEndpoints(page playwright.Page, baseURL string) ([]string, error) {
	var hiddenEndpoints []string

	// Find all <script src="..."></script>
	scripts, err := page.Locator("script[src]").EvaluateAll(`elements => elements.map(e => e.src)`)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	// Broad regex to catch common API patterns encoded in javascript strings
	// Looks for strings starting with /api/, /v1/, /v2/, /internal/
	apiRegex := regexp.MustCompile(`"(/(?:api|v[1-9]|internal)/[^"\s]+)"|'((?:/api|v[1-9]|internal)/[^'\s]+)'`)

	if scriptURLs, ok := scripts.([]interface{}); ok {
		for _, s := range scriptURLs {
			if strUrl, ok := s.(string); ok {
				// Only fetch JS files from our target domain to avoid crawling 3rd party CDNs
				parsed, err := url.Parse(strUrl)
				if err != nil || parsed.Host != base.Host {
					continue
				}

				absURL := base.ResolveReference(parsed).String()

				// Fetch the raw JS content using Playwright's evaluation to bypass CORS
				jsContent, err := page.Evaluate(fmt.Sprintf(`async () => {
					try {
						let res = await fetch("%s");
						return await res.text();
					} catch (e) { return ""; }
				}`, absURL))

				if err != nil {
					continue
				}

				if contentStr, ok := jsContent.(string); ok && contentStr != "" {
					matches := apiRegex.FindAllStringSubmatch(contentStr, -1)
					for _, match := range matches {
						// match[1] corresponds to double quotes, match[2] to single quotes
						endpoint := match[1]
						if endpoint == "" {
							endpoint = match[2]
						}

						if endpoint != "" {
							// Check if it's already a full URL or just a path
							if !strings.HasPrefix(endpoint, "http") {
								endpointURL, err := url.Parse(endpoint)
								if err == nil {
									endpoint = base.ResolveReference(endpointURL).String()
								}
							}
							hiddenEndpoints = append(hiddenEndpoints, endpoint)
						}
					}
				}
			}
		}
	}

	return hiddenEndpoints, nil
}
