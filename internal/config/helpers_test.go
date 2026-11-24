package config

import (
	"reflect"
	"testing"
)

// TestParseDelimitedList verifies values are split, trimmed and empty
// entries removed.
func TestParseDelimitedList(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "trim and drop empty", in: " topic1 ; topic2 ;; topic3 ", want: []string{"topic1", "topic2", "topic3"}},
		{name: "single value", in: "origin/a/wis2/#", want: []string{"origin/a/wis2/#"}},
		{name: "empty", in: "", want: nil},
		{name: "only separators and spaces", in: " ; ;; ", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDelimitedList(tt.in, ";"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDelimitedList(%q) = %#v; want %#v", tt.in, got, tt.want)
			}
		})
	}
}
