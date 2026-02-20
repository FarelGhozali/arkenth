package crawler

import (
	"log"
	"net/url"

	"web-qa-automation/config"
	"web-qa-automation/reporter"
	"web-qa-automation/tester"

	"github.com/playwright-community/playwright-go"
)

type Crawler struct {
	Config  *config.Config
	Report  *reporter.Report
	Browser playwright.Browser
	Visited map[string]bool
}

func NewCrawler(cfg *config.Config, rep *reporter.Report, browser playwright.Browser) *Crawler {
	return &Crawler{
		Config:  cfg,
		Report:  rep,
		Browser: browser,
		Visited: make(map[string]bool),
	}
}

func (c *Crawler) Start() error {
	return c.crawl(c.Config.TargetURL, 0)
}

func (c *Crawler) crawl(targetURL string, currentDepth int) error {
	if currentDepth > c.Config.Depth {
		return nil
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return err
	}
	parsedURL.Fragment = ""
	cleanURL := parsedURL.String()

	if c.Visited[cleanURL] {
		return nil
	}
	c.Visited[cleanURL] = true
	c.Report.AddScannedURL(cleanURL)

	log.Printf("Crawling: %s (Depth: %d)", cleanURL, currentDepth)

	// Create a new browser context for each unique URL crawl to isolate state somewhat
	// IgnoreHttpsErrors is added so localhost with self-signed certs won't crash
	context, err := c.Browser.NewContext(playwright.BrowserNewContextOptions{
		IgnoreHttpsErrors: playwright.Bool(true),
	})
	if err != nil {
		return err
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return err
	}

	tester.SetupPassiveMonitors(page, c.Report)
	tester.MonitorPageError(page, c.Report)

	// Navigate to target
	_, err = page.Goto(cleanURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	if err != nil {
		log.Printf("Failed to navigate to %s: %v", cleanURL, err)
		return nil
	}

	// Active QA / Fuzzing Phase
	tester.FuzzPage(page, c.Report, cleanURL)

	// Re-navigate or restore the context if fuzzing caused navigation / crash
	// We'll extract links using a fresh page to avoid mutated DOM issues
	pageContext, err := context.NewPage()
	if err == nil {
		pageContext.Goto(cleanURL, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		if currentDepth < c.Config.Depth {
			links, err := extractInternalLinks(pageContext, cleanURL)
			if err != nil {
				log.Printf("Failed to extract links from %s: %v", cleanURL, err)
			} else {
				for _, link := range links {
					c.crawl(link, currentDepth+1)
				}
			}
		}
		pageContext.Close()
	}

	return nil
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

					// Ignore entirely invalid protocols like javascript: or mailto:
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
