package mqtt

import "testing"

// TestPayloadStructure verifies that Payload struct can be
// correctly instantiated and holds expected fields.
func TestPayloadStrcuture(t *testing.T) {

	p := Payload{
		Properties: Properties{
			DataID: "123",
		},
		Links: []Link{
			{
				Href: "http://example.com/file.bufr",
				Rel:  "canonical",
				Type: "application/octet-stream",
			},
		},
	}

	if p.Properties.DataID != "123" {
		t.Errorf("expected DataID 123, got %s", p.Properties.DataID)
	}

	if len(p.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(p.Links))
	}

	if p.Links[0].Href == "" {
		t.Errorf("expected valid href")
	}
}
