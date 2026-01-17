package metadata

import (
	"fmt"
	"strconv"
)

// parseFloat64 parses a string to float64.
func parseFloat64(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty value")
	}

	return strconv.ParseFloat(s, 64)
}

// parseInt parses a string to int.
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty value")
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, fmt.Errorf("value must be non-negative, got %d", i)
	}

	return i, nil
}

// parseBool parses a string to bool pointer.
// If value is empty, returns true (flag without value means true).
// If value is "true" or "false", parses accordingly.
// Otherwise returns nil.
func parseBool(value string) *bool {
	if value == "" {
		b := true

		return &b
	}
	if value == "true" {
		b := true

		return &b
	}
	if value == "false" {
		b := false

		return &b
	}
	// Invalid value, return nil

	return nil
}
