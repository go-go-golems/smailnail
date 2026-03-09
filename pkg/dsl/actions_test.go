package dsl

import (
	"testing"

	"github.com/emersion/go-imap/v2"
)

func TestBuildUIDSetUsesUIDs(t *testing.T) {
	messages := []*EmailMessage{
		{UID: 42},
		{UID: 77},
	}

	uidSet := buildUIDSet(messages)

	if !uidSet.Contains(imap.UID(42)) {
		t.Fatalf("expected uid set to contain 42")
	}
	if !uidSet.Contains(imap.UID(77)) {
		t.Fatalf("expected uid set to contain 77")
	}

	nums, ok := uidSet.Nums()
	if !ok {
		t.Fatalf("expected uid set to have concrete numbers")
	}
	if len(nums) != 2 {
		t.Fatalf("expected 2 UIDs, got %d", len(nums))
	}
}
