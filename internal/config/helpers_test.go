package config

import (
	"reflect"
	"testing"
)

// TestParseDelimitedList verifies that values are split,
// trimmed and empty entries removed.
func TestParseDelimitedList(t *testing.T) {

	result := parseDelimitedList(
		" topic1 ; topic2 ;; topic3 ",
		";",
	)

	expected := []string{
		"topic1",
		"topic2",
		"topic3",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf(
			"expected %v got %v",
			expected,
			result,
		)
	}
}

// TestParseDelimitedListEmpty verifies that an empty input
// produces an empty slice.
func TestParseDelimitedListEmpty(t *testing.T) {

	result := parseDelimitedList("", ";")

	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %v", result)
	}
}
