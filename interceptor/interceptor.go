package interceptor

import (
	"fmt"
	"log"
	"strings"

	"github.com/FarelGhozali/arkenth/config"
	"github.com/FarelGhozali/arkenth/models"

	"github.com/playwright-community/playwright-go"
)

type Interceptor struct {
	Config    *config.AppConfig
	Report    *models.Report
	Anomalies []models.NetworkAnomaly
}

func NewInterceptor(cfg *config.AppConfig, rep *models.Report) *Interceptor {
	return &Interceptor{
		Config:    cfg,
		Report:    rep,
		Anomalies: []models.NetworkAnomaly{},
	}
}

// AttachToPage sets up request/response monitoring and route blocking (FastMode)
func (i *Interceptor) AttachToPage(page playwright.Page) {
	if i.Config.FastMode {
		log.Println("Fast Mode enabled: Blocking image, media, font, and tracker requests.")
		page.Route("**/*", func(route playwright.Route) {
			req := route.Request()
			rt := req.ResourceType()

			// Block heavy assets
			if rt == "image" || rt == "media" || rt == "font" {
				route.Abort()
				return
			}

			// Simple heuristic to block ad/tracking scripts
			url := req.URL()
			if strings.Contains(url, "google-analytics.com") ||
				strings.Contains(url, "doubleclick.net") ||
				strings.Contains(url, "googletagmanager.com") ||
				strings.Contains(url, "facebook.net") {
				route.Abort()
				return
			}

			route.Continue()
		})
	}

	// Intercept and Log HTTP Status Codes >= 400
	page.OnResponse(func(response playwright.Response) {
		status := response.Status()
		if status >= 400 {
			req := response.Request()
			payloadData, _ := req.PostData()

			// [Phase 2: Advanced] Server & WAF Fingerprinting
			headers := response.Headers()
			var fingerprint strings.Builder

			// Detect WAF
			if _, ok := headers["cf-ray"]; ok {
				fingerprint.WriteString("[WAF: Cloudflare] ")
			} else if _, ok := headers["x-amz-cf-id"]; ok {
				fingerprint.WriteString("[WAF: AWS CloudFront] ")
			} else if _, ok := headers["x-sucuri-id"]; ok {
				fingerprint.WriteString("[WAF: Sucuri] ")
			}

			// Detect Backend/Server
			if server, ok := headers["server"]; ok {
				fingerprint.WriteString(fmt.Sprintf("[Server: %s] ", server))
			}
			if poweredBy, ok := headers["x-powered-by"]; ok {
				fingerprint.WriteString(fmt.Sprintf("[Framework: %s] ", poweredBy))
			}

			fpString := strings.TrimSpace(fingerprint.String())

			errorDesc := response.StatusText()
			if fpString != "" {
				errorDesc = fmt.Sprintf("%s | Intel: %s", errorDesc, fpString)
				log.Printf("🕵️  Fingerprint found on %s: %s", req.URL(), fpString)
			}

			anomaly := models.NetworkAnomaly{
				URL:      req.URL(),
				Method:   req.Method(),
				Status:   status,
				ErrorMsg: errorDesc,
				Payload:  payloadData,
			}
			i.Anomalies = append(i.Anomalies, anomaly)
			i.Report.NetworkAnomalies = append(i.Report.NetworkAnomalies, anomaly)
		}
	})

	// Intercept Request Failures (Timeouts, DNS issues)
	page.OnRequestFailed(func(request playwright.Request) {
		errText := ""
		if request.Failure() != nil {
			errText = request.Failure().Error()
		}

		anomaly := models.NetworkAnomaly{
			URL:      request.URL(),
			Method:   request.Method(),
			Status:   0, // TCP/DNS failure, no HTTP status reached
			ErrorMsg: errText,
		}
		i.Anomalies = append(i.Anomalies, anomaly)
		i.Report.NetworkAnomalies = append(i.Report.NetworkAnomalies, anomaly)
	})
}
