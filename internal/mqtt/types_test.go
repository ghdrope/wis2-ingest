package mqtt

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestPayloadUnmarshal verifies a WIS2 notification message is decoded with
// the expected JSON field names, ignoring unknown fields.
func TestPayloadUnmarshal(t *testing.T) {
	raw := `{
		"id": "31e9d66a-cd83-4174-9429-b932f1abe1be",
		"type": "Feature",
		"properties": {
			"data_id": "fr-meteofrance/data/core/weather/surface-based-observations/synop/A_ISMD01LFPW",
			"pubtime": "2026-01-01T00:00:00Z"
		},
		"links": [
			{"href": "https://example.org/file.bufr4", "rel": "canonical", "type": "application/octet-stream", "length": 123},
			{"href": "https://example.org/meta", "rel": "via", "type": "text/html"}
		]
	}`

	var p Payload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	want := Payload{
		Properties: Properties{DataID: "fr-meteofrance/data/core/weather/surface-based-observations/synop/A_ISMD01LFPW"},
		Links: []Link{
			{Href: "https://example.org/file.bufr4", Rel: "canonical", Type: "application/octet-stream"},
			{Href: "https://example.org/meta", Rel: "via", Type: "text/html"},
		},
	}
	if !reflect.DeepEqual(p, want) {
		t.Errorf("Payload =\n%+v\nwant\n%+v", p, want)
	}
}
