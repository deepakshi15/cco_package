package ConvertData

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
)

func ConvertNetwork(value string) string {
	// Trim spaces and convert to lower case for uniformity
	value = strings.TrimSpace(strings.ToLower(value))

	// Regular expression to extract the numeric part and the unit
	re := regexp.MustCompile(`(\d+\.?\d*)(\s*(gigabit|megabit|gb|mb))?`)
	matches := re.FindStringSubmatch(value)

	if len(matches) < 3 {
		return "Invalid"
	}

	// Extract the numeric value
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "Invalid"
	}

	// Extract unit (Gigabit or Megabit)
	unit := matches[2]

	// Convert units to Megabit (1 Gigabit = 1000 Megabits)
	switch {
	case strings.Contains(unit, "gigabit") || strings.Contains(unit, "gb"):
		num *= 1000 // Convert Gigabit to Megabit
	case strings.Contains(unit, "megabit") || strings.Contains(unit, "mb"):
		// No conversion needed, already in Megabit
	default:
		return "Invalid"
	}

	// Return the value in Megabits as a string
	return fmt.Sprintf("%.0f", num)
}
func ConvertMemory(value string) string {
	// Trim spaces and convert to lower case for uniformity
	value = strings.TrimSpace(strings.ToLower(value))

	// Regular expression to extract the numeric part and the unit (e.g., GiB)
	re := regexp.MustCompile(`(\d+\.?\d*)(\s*(gib))?`)
	matches := re.FindStringSubmatch(value)

	if len(matches) < 2 {
		return "Invalid"
	}

	// Extract the numeric value
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "Invalid"
	}

	// Return the value as a string (GiB values are treated as numeric without conversion)
	return fmt.Sprintf("%.0f", num)
}
func ConvertYear(value string) string {
	// Trim spaces and convert to lower case for uniformity
	value = strings.TrimSpace(strings.ToLower(value))

	// Regular expression to extract the numeric part and the unit (e.g., yr)
	re := regexp.MustCompile(`(\d+\.?\d*)(\s*(yr))?`)
	matches := re.FindStringSubmatch(value)

	if len(matches) < 2 {
		return "Invalid"
	}

	// Extract the numeric value
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "Invalid"
	}

	// Return the value as a string (yr values are treated as numeric without conversion)
	return fmt.Sprintf("%.0f", num)
}


