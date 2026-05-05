package swagger

// EndpointParam represents a single parameter for an API endpoint.
type EndpointParam struct {
	Name     string // Parameter name (e.g., "id", "email")
	In       string // Location: "path", "query", "header", "cookie"
	Type     string // Data type: "string", "integer", "number", "boolean", "array", "object"
	Format   string // Format hint: "email", "date-time", "uuid", "int32", "int64", etc.
	Required bool   // Whether this parameter is required
}

// BodyField represents a single field inside a JSON request body schema.
type BodyField struct {
	Name     string // Field name (e.g., "price", "email")
	Type     string // Data type: "string", "integer", "number", "boolean", "array", "object"
	Format   string // Format hint: "email", "date-time", "uuid", etc.
	Required bool   // Whether this field is required by the schema
}

// Endpoint represents a single API endpoint extracted from an OpenAPI spec.
type Endpoint struct {
	Path        string          // URL path (e.g., "/api/users/{id}")
	Method      string          // HTTP method (GET, POST, PUT, DELETE, PATCH)
	OperationID string          // Unique operation identifier from the spec
	Summary     string          // Human-readable summary of the operation
	Params      []EndpointParam // Path, query, header, cookie parameters
	BodyFields  []BodyField     // JSON body schema fields (for POST/PUT/PATCH)
	ContentType string          // Request body content type (e.g., "application/json")
}

// FuzzResult represents the outcome of a single fuzz test against an endpoint.
type FuzzResult struct {
	Endpoint     string `json:"endpoint"`      // Full URL that was tested
	Method       string `json:"method"`        // HTTP method used
	Payload      string `json:"payload"`       // The fuzz payload that was sent
	FieldName    string `json:"field_name"`    // Which field/parameter was fuzzed
	StatusCode   int    `json:"status_code"`   // HTTP response status code
	ResponseBody string `json:"response_body"` // Truncated response body for analysis
	Severity     string `json:"severity"`      // "CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"
	Category     string `json:"category"`      // "crash", "sqli", "xss", "type_confusion", "boundary", etc.
	Description  string `json:"description"`   // Human-readable description of the finding
	IsAnomaly    bool   `json:"is_anomaly"`    // Whether this result is considered an anomaly
}

// FuzzReport is the top-level report aggregating all fuzzing results.
type FuzzReport struct {
	SpecURL         string       `json:"spec_url"`         // URL or path to the OpenAPI spec
	BaseURL         string       `json:"base_url"`         // Base URL of the target API
	TotalEndpoints  int          `json:"total_endpoints"`  // Number of endpoints discovered
	TotalRequests   int          `json:"total_requests"`   // Total fuzz requests sent
	TotalAnomalies  int          `json:"total_anomalies"`  // Total anomalies detected
	CriticalCount   int          `json:"critical_count"`   // Count of CRITICAL severity findings
	HighCount       int          `json:"high_count"`       // Count of HIGH severity findings
	MediumCount     int          `json:"medium_count"`     // Count of MEDIUM severity findings
	Endpoints       []Endpoint   `json:"endpoints"`        // All discovered endpoints
	Results         []FuzzResult `json:"results"`          // All fuzz results
	Anomalies       []FuzzResult `json:"anomalies"`        // Only anomalous results
	DurationSeconds float64      `json:"duration_seconds"` // Total fuzzing duration
}
