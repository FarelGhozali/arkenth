package swagger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Fuzzer orchestrates the API fuzzing process using parsed endpoints.
type Fuzzer struct {
	BaseURL     string
	Endpoints   []Endpoint
	Report      *FuzzReport
	Concurrency int
	Client      *http.Client
	mu          sync.Mutex
	totalSent   int64
}

// NewFuzzer creates a new API Fuzzer instance.
func NewFuzzer(baseURL string, endpoints []Endpoint, concurrency int) *Fuzzer {
	if concurrency <= 0 {
		concurrency = 10
	}

	return &Fuzzer{
		BaseURL:     strings.TrimRight(baseURL, "/"),
		Endpoints:   endpoints,
		Concurrency: concurrency,
		Report: &FuzzReport{
			BaseURL:        baseURL,
			TotalEndpoints: len(endpoints),
			Endpoints:      endpoints,
		},
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: concurrency * 2,
				IdleConnTimeout:     30 * time.Second,
			},
		},
	}
}

// Run executes the full fuzzing campaign against all endpoints.
func (f *Fuzzer) Run() *FuzzReport {
	start := time.Now()

	log.Printf("🚀 Starting API Fuzzing: %d endpoints @ %s (concurrency: %d)",
		len(f.Endpoints), f.BaseURL, f.Concurrency)

	// Build a work queue of all fuzz jobs
	jobs := f.buildJobs()
	log.Printf("📦 Generated %d fuzz jobs across all endpoints", len(jobs))

	// Process jobs concurrently using a worker pool
	jobCh := make(chan fuzzJob, len(jobs))
	var wg sync.WaitGroup

	// Launch workers
	for i := 0; i < f.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				f.executeJob(job)
			}
		}()
	}

	// Feed jobs
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)

	wg.Wait()

	f.Report.DurationSeconds = time.Since(start).Seconds()
	f.Report.TotalRequests = int(atomic.LoadInt64(&f.totalSent))

	// Separate anomalies
	for _, r := range f.Report.Results {
		if r.IsAnomaly {
			f.Report.Anomalies = append(f.Report.Anomalies, r)
			switch r.Severity {
			case "CRITICAL":
				f.Report.CriticalCount++
			case "HIGH":
				f.Report.HighCount++
			case "MEDIUM":
				f.Report.MediumCount++
			}
		}
	}
	f.Report.TotalAnomalies = len(f.Report.Anomalies)

	log.Printf("✅ API Fuzzing complete in %.1fs: %d requests, %d anomalies (%d critical, %d high, %d medium)",
		f.Report.DurationSeconds, f.Report.TotalRequests, f.Report.TotalAnomalies,
		f.Report.CriticalCount, f.Report.HighCount, f.Report.MediumCount)

	return f.Report
}

// fuzzJob represents a single fuzz request to execute.
type fuzzJob struct {
	Endpoint Endpoint
	Field    string // Which field/param is being fuzzed
	Payload  string
	Category string // Payload category name
	Severity string // Expected severity if anomaly found
}

// buildJobs generates all fuzz jobs for all endpoints.
func (f *Fuzzer) buildJobs() []fuzzJob {
	var jobs []fuzzJob

	for _, ep := range f.Endpoints {
		// Fuzz path parameters
		for _, param := range ep.Params {
			categories := GetPayloadsForType(param.Type, param.Format)
			for _, cat := range categories {
				for _, payload := range cat.Payloads {
					jobs = append(jobs, fuzzJob{
						Endpoint: ep,
						Field:    fmt.Sprintf("%s:%s", param.In, param.Name),
						Payload:  payload,
						Category: cat.Name,
						Severity: cat.Severity,
					})
				}
			}
		}

		// Fuzz body fields
		for _, field := range ep.BodyFields {
			categories := GetPayloadsForType(field.Type, field.Format)
			for _, cat := range categories {
				for _, payload := range cat.Payloads {
					jobs = append(jobs, fuzzJob{
						Endpoint: ep,
						Field:    fmt.Sprintf("body:%s", field.Name),
						Payload:  payload,
						Category: cat.Name,
						Severity: cat.Severity,
					})
				}
			}
		}

		// If no params/body, still fuzz the endpoint with security payloads on the URL itself
		if len(ep.Params) == 0 && len(ep.BodyFields) == 0 {
			for _, cat := range securityPayloads() {
				for _, payload := range cat.Payloads {
					jobs = append(jobs, fuzzJob{
						Endpoint: ep,
						Field:    "url",
						Payload:  payload,
						Category: cat.Name,
						Severity: cat.Severity,
					})
				}
			}
		}
	}

	return jobs
}

// executeJob runs a single fuzz request and records the result.
func (f *Fuzzer) executeJob(job fuzzJob) {
	atomic.AddInt64(&f.totalSent, 1)

	// Build the request URL
	fullURL := f.buildURL(job)

	// Build HTTP request
	var req *http.Request
	var err error

	if strings.HasPrefix(job.Field, "body:") {
		// Build a JSON body with the payload injected into the target field
		body := f.buildFuzzBody(job)
		req, err = http.NewRequest(job.Endpoint.Method, fullURL, bytes.NewBuffer(body))
		if req != nil {
			req.Header.Set("Content-Type", "application/json")
		}
	} else {
		req, err = http.NewRequest(job.Endpoint.Method, fullURL, nil)
	}

	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "WebQA-SwaggerFuzzer/1.0")

	resp, err := f.Client.Do(req)
	if err != nil {
		// Network error — record as anomaly
		result := FuzzResult{
			Endpoint:    fullURL,
			Method:      job.Endpoint.Method,
			Payload:     truncate(job.Payload, 200),
			FieldName:   job.Field,
			StatusCode:  0,
			Severity:    "MEDIUM",
			Category:    job.Category,
			Description: fmt.Sprintf("Network error: %v", err),
			IsAnomaly:   true,
		}
		f.addResult(result)
		return
	}
	defer resp.Body.Close()

	// Read response body (truncated)
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	respBody := string(bodyBytes)

	// Analyze response for anomalies
	result := FuzzResult{
		Endpoint:     fullURL,
		Method:       job.Endpoint.Method,
		Payload:      truncate(job.Payload, 200),
		FieldName:    job.Field,
		StatusCode:   resp.StatusCode,
		ResponseBody: truncate(respBody, 500),
		Category:     job.Category,
	}

	// Determine if this is an anomaly
	f.classifyResult(&result, job)

	f.addResult(result)
}

// buildURL constructs the full URL, injecting payloads into path parameters.
func (f *Fuzzer) buildURL(job fuzzJob) string {
	path := job.Endpoint.Path

	// Replace path parameters with payload if targeting a path param
	if strings.HasPrefix(job.Field, "path:") {
		paramName := strings.TrimPrefix(job.Field, "path:")
		path = strings.ReplaceAll(path, "{"+paramName+"}", job.Payload)
	}

	// Replace any remaining {param} placeholders with a safe default
	for _, param := range job.Endpoint.Params {
		if param.In == "path" {
			placeholder := "{" + param.Name + "}"
			if strings.Contains(path, placeholder) {
				path = strings.ReplaceAll(path, placeholder, "1")
			}
		}
	}

	fullURL := f.BaseURL + path

	// Add query parameters
	if strings.HasPrefix(job.Field, "query:") {
		paramName := strings.TrimPrefix(job.Field, "query:")
		if strings.Contains(fullURL, "?") {
			fullURL += "&" + paramName + "=" + job.Payload
		} else {
			fullURL += "?" + paramName + "=" + job.Payload
		}
	}

	return fullURL
}

// buildFuzzBody creates a JSON request body with the payload injected into the target field.
func (f *Fuzzer) buildFuzzBody(job fuzzJob) []byte {
	fieldName := strings.TrimPrefix(job.Field, "body:")

	body := make(map[string]interface{})

	// Fill all body fields with valid defaults, except the one being fuzzed
	for _, field := range job.Endpoint.BodyFields {
		if field.Name == fieldName {
			body[field.Name] = job.Payload
		} else {
			body[field.Name] = getDefaultValue(field.Type)
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return []byte(fmt.Sprintf(`{"%s": "%s"}`, fieldName, job.Payload))
	}
	return data
}

// classifyResult analyzes the HTTP response and determines if it's an anomaly.
func (f *Fuzzer) classifyResult(result *FuzzResult, job fuzzJob) {
	// HTTP 5xx = Server crash (always critical)
	if result.StatusCode >= 500 {
		result.IsAnomaly = true
		result.Severity = "CRITICAL"
		result.Description = fmt.Sprintf("Server crash (HTTP %d) triggered by %s payload on field '%s'",
			result.StatusCode, job.Category, job.Field)
		return
	}

	// Check response body for sensitive error indicators
	bodyLower := strings.ToLower(result.ResponseBody)

	// Stack trace / debug info leakage
	stackTraceIndicators := []string{
		"stack trace", "traceback", "at line", "exception in",
		"panic:", "goroutine", "runtime error",
		"sqlstate", "sql syntax", "mysql", "postgresql", "sqlite",
		"ora-", "microsoft ole db",
	}
	for _, indicator := range stackTraceIndicators {
		if strings.Contains(bodyLower, indicator) {
			result.IsAnomaly = true
			result.Severity = "HIGH"
			result.Description = fmt.Sprintf("Sensitive error information leaked (contains '%s') from %s payload",
				indicator, job.Category)
			return
		}
	}

	// Reflected XSS detection
	if job.Category == "xss" && strings.Contains(result.ResponseBody, job.Payload) {
		result.IsAnomaly = true
		result.Severity = "HIGH"
		result.Description = fmt.Sprintf("Potential reflected XSS — payload was echoed back in response body")
		return
	}

	// Not an anomaly
	result.IsAnomaly = false
	result.Severity = "INFO"
	result.Description = fmt.Sprintf("Normal response (HTTP %d)", result.StatusCode)
}

// addResult safely appends a result to the report.
func (f *Fuzzer) addResult(result FuzzResult) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Report.Results = append(f.Report.Results, result)
}

// getDefaultValue returns a safe default value for a given data type.
func getDefaultValue(dataType string) interface{} {
	switch dataType {
	case "integer":
		return 1
	case "number":
		return 1.0
	case "boolean":
		return true
	case "array":
		return []interface{}{}
	default:
		return "test"
	}
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
