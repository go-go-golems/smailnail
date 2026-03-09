package dsl

import "testing"

func TestDetermineRequiredBodySectionsWithoutMimePartsDoesNotNeedStructure(t *testing.T) {
	config := OutputConfig{
		Fields: []interface{}{
			Field{Name: "uid"},
			Field{Name: "subject"},
		},
	}

	parts, err := determineRequiredBodySections(nil, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(parts) != 0 {
		t.Fatalf("expected no MIME parts, got %d", len(parts))
	}
}
