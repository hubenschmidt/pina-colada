package lib

import (
	"net/http"
	"strconv"
)

// ParseIntQueryParam parses an integer query parameter with a default value.
func ParseIntQueryParam(r *http.Request, name string, defaultVal int) int {
	v := r.URL.Query().Get(name)
	if v == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return parsed
}

// ParseOptionalInt64Param parses an optional int64 query parameter.
// Returns (nil, nil) if empty, (*value, nil) on success, (nil, error) on parse failure.
func ParseOptionalInt64Param(r *http.Request, name string) (*int64, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
