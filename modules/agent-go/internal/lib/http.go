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
