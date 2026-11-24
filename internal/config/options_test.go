package config

import (
	"reflect"
	"testing"
)

// TestNewIngestOptions verifies the defaults: the public WIS2 global broker
// with its anonymous credentials, and nothing else set.
func TestNewIngestOptions(t *testing.T) {
	want := &IngestOptions{
		Host:     "globalbroker.meteo.fr",
		Username: "everyone",
		Password: "everyone",
	}
	if got := NewIngestOptions(); !reflect.DeepEqual(got, want) {
		t.Errorf("NewIngestOptions() = %+v; want %+v", got, want)
	}
}
