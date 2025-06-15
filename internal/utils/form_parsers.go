package utils

import (
	"strconv"
	"time"
)

// ParseDate parses a date string in YYYY-MM-DD format
func ParseDate(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ParseInt64 parses a string to int64
func ParseInt64(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// ParseFloat64 parses a string to float64
func ParseFloat64(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return f
}
