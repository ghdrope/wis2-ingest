package config

import "strings"

// parseDelimitedList converts a delimited string into a slice,
// trimming whitespace and discarding empty entries.
func parseDelimitedList(s, sep string) []string {
	parts := strings.Split(s, sep)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
