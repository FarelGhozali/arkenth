package swagger

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

// Parser loads and parses OpenAPI/Swagger specifications.
type Parser struct {
	Spec *openapi3.T
}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// LoadSpec loads an OpenAPI spec from a URL or local file path.
// Supports both Swagger 2.0 and OpenAPI 3.x formats.
func (p *Parser) LoadSpec(urlOrPath string) error {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	var err error

	// Determine if the input is a URL or a file path
	u, parseErr := url.Parse(urlOrPath)
	if parseErr == nil && (u.Scheme == "http" || u.Scheme == "https") {
		p.Spec, err = loader.LoadFromURI(u)
	} else {
		p.Spec, err = loader.LoadFromFile(urlOrPath)
	}

	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the spec (non-fatal warnings)
	if valErr := p.Spec.Validate(context.Background()); valErr != nil {
		log.Printf("⚠️  OpenAPI spec validation warnings: %v", valErr)
	}

	log.Printf("✅ OpenAPI spec loaded successfully: %s (version: %s)", p.getTitle(), p.Spec.OpenAPI)
	return nil
}

// ExtractEndpoints parses the loaded spec and returns a list of Endpoint structs
// with all parameters, body fields, and metadata extracted.
func (p *Parser) ExtractEndpoints() ([]Endpoint, error) {
	if p.Spec == nil {
		return nil, fmt.Errorf("spec not loaded, call LoadSpec first")
	}

	if p.Spec.Paths == nil {
		return nil, fmt.Errorf("no paths found in the OpenAPI spec")
	}

	var endpoints []Endpoint

	for path, pathItem := range p.Spec.Paths.Map() {
		for method, op := range pathItem.Operations() {
			if op == nil {
				continue
			}

			ep := Endpoint{
				Path:        path,
				Method:      method,
				OperationID: op.OperationID,
				Summary:     op.Summary,
			}

			// Extract parameters (path, query, header, cookie)
			ep.Params = p.extractParams(op)

			// Extract request body fields (for POST/PUT/PATCH)
			ep.BodyFields, ep.ContentType = p.extractBodyFields(op)

			endpoints = append(endpoints, ep)
		}
	}

	log.Printf("📋 Extracted %d endpoints from spec", len(endpoints))
	return endpoints, nil
}

// extractParams extracts all parameters from an operation.
func (p *Parser) extractParams(op *openapi3.Operation) []EndpointParam {
	var params []EndpointParam

	for _, paramRef := range op.Parameters {
		if paramRef == nil || paramRef.Value == nil {
			continue
		}

		param := paramRef.Value
		ep := EndpointParam{
			Name:     param.Name,
			In:       param.In,
			Required: param.Required,
		}

		if param.Schema != nil && param.Schema.Value != nil {
			schema := param.Schema.Value
			ep.Type = p.getSchemaType(schema)
			ep.Format = schema.Format
		}

		params = append(params, ep)
	}

	return params
}

// extractBodyFields extracts request body field definitions from an operation.
func (p *Parser) extractBodyFields(op *openapi3.Operation) ([]BodyField, string) {
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return nil, ""
	}

	content := op.RequestBody.Value.Content

	// Prefer application/json
	for contentType, mediaType := range content {
		if mediaType.Schema == nil || mediaType.Schema.Value == nil {
			continue
		}

		schema := mediaType.Schema.Value
		fields := p.extractFieldsFromSchema(schema, op.RequestBody.Value.Required)
		return fields, contentType
	}

	return nil, ""
}

// extractFieldsFromSchema recursively extracts fields from an OpenAPI schema.
func (p *Parser) extractFieldsFromSchema(schema *openapi3.Schema, bodyRequired bool) []BodyField {
	var fields []BodyField

	if schema.Type == nil {
		return fields
	}

	if schema.Type.Is("object") {
		requiredSet := make(map[string]bool)
		for _, r := range schema.Required {
			requiredSet[r] = true
		}

		for name, propRef := range schema.Properties {
			if propRef == nil || propRef.Value == nil {
				continue
			}

			prop := propRef.Value
			fields = append(fields, BodyField{
				Name:     name,
				Type:     p.getSchemaType(prop),
				Format:   prop.Format,
				Required: requiredSet[name],
			})
		}
	}

	return fields
}

// getSchemaType extracts the type string from an OpenAPI schema.
func (p *Parser) getSchemaType(schema *openapi3.Schema) string {
	if schema.Type == nil {
		return "string" // Default fallback
	}

	// kin-openapi v0.128+ uses Types (slice), check for single type
	types := schema.Type.Slice()
	if len(types) > 0 {
		return types[0]
	}

	return "string"
}

// getTitle returns the spec title or a fallback.
func (p *Parser) getTitle() string {
	if p.Spec.Info != nil && p.Spec.Info.Title != "" {
		return p.Spec.Info.Title
	}
	return "Untitled API"
}
