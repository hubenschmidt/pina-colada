package lib

import "time"

// ParseDateString parses a date string, returning nil if nil input or parse error.
func ParseDateString(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}
