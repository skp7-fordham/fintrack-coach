package importcsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

var requiredHeaders = []string{"date", "description", "amount", "type"}
var optionalHeaders = []string{"merchant", "category", "notes"}

type ParseResult struct {
	Rows []domain.ParsedImportRow
}

func Parse(reader io.Reader, maxRows int) (*ParseResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true
	csvReader.ReuseRecord = false

	headerRow, err := csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("%w: missing header row", domain.ErrInvalidCSV)
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidCSV, err)
	}

	index, err := mapHeaders(headerRow)
	if err != nil {
		return nil, err
	}

	result := &ParseResult{Rows: make([]domain.ParsedImportRow, 0)}
	rowNumber := 1 // header is row 1; data starts at 2

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		rowNumber++
		if err != nil {
			return nil, fmt.Errorf("%w: malformed row %d", domain.ErrInvalidCSV, rowNumber)
		}

		if isBlankRecord(record) {
			continue
		}

		if len(result.Rows) >= maxRows {
			return nil, fmt.Errorf("%w: exceeds maximum of %d rows", domain.ErrInvalidCSV, maxRows)
		}

		get := func(name string) string {
			idx, ok := index[name]
			if !ok || idx >= len(record) {
				return ""
			}
			value := strings.TrimSpace(record[idx])
			switch name {
			case "description", "merchant", "category", "notes":
				return neutralizeFormula(value)
			default:
				return value
			}
		}

		raw := map[string]string{
			"date":        get("date"),
			"description": get("description"),
			"merchant":    get("merchant"),
			"amount":      get("amount"),
			"type":        get("type"),
			"category":    get("category"),
			"notes":       get("notes"),
		}

		result.Rows = append(result.Rows, domain.ParsedImportRow{
			RowNumber:   rowNumber,
			Date:        raw["date"],
			Description: raw["description"],
			Merchant:    raw["merchant"],
			Amount:      raw["amount"],
			Type:        raw["type"],
			Category:    raw["category"],
			Notes:       raw["notes"],
			Raw:         raw,
		})
	}

	return result, nil
}

func mapHeaders(headers []string) (map[string]int, error) {
	index := make(map[string]int, len(headers))
	seen := make(map[string]struct{}, len(headers))

	for i, header := range headers {
		name := strings.ToLower(strings.TrimSpace(header))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("%w: duplicate header %q", domain.ErrInvalidCSV, name)
		}
		seen[name] = struct{}{}
		index[name] = i
	}

	for _, required := range requiredHeaders {
		if _, ok := index[required]; !ok {
			return nil, fmt.Errorf("%w: missing required header %q", domain.ErrInvalidCSV, required)
		}
	}

	allowed := make(map[string]struct{}, len(requiredHeaders)+len(optionalHeaders))
	for _, h := range requiredHeaders {
		allowed[h] = struct{}{}
	}
	for _, h := range optionalHeaders {
		allowed[h] = struct{}{}
	}
	for name := range index {
		if _, ok := allowed[name]; !ok {
			return nil, fmt.Errorf("%w: unexpected header %q", domain.ErrInvalidCSV, name)
		}
	}

	return index, nil
}

func isBlankRecord(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

func neutralizeFormula(value string) string {
	if value == "" {
		return value
	}
	if strings.HasPrefix(value, "=") ||
		strings.HasPrefix(value, "+") ||
		strings.HasPrefix(value, "-") ||
		strings.HasPrefix(value, "@") {
		return "'" + value
	}
	return value
}
