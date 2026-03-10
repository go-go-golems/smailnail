package mailutil

import (
	"testing"

	"github.com/emersion/go-message/mail"
)

func TestSetSingleAddressParsesDisplayName(t *testing.T) {
	header := mail.Header{}

	err := SetSingleAddress(&header, "From", "John Doe <john@example.com>")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	addresses, err := header.AddressList("From")
	if err != nil {
		t.Fatalf("expected to parse serialized header, got %v", err)
	}
	if len(addresses) != 1 {
		t.Fatalf("expected 1 address, got %d", len(addresses))
	}
	if addresses[0].Name != "John Doe" {
		t.Fatalf("expected display name John Doe, got %q", addresses[0].Name)
	}
	if addresses[0].Address != "john@example.com" {
		t.Fatalf("expected mailbox john@example.com, got %q", addresses[0].Address)
	}
}

func TestSetSingleAddressRejectsInvalidAddress(t *testing.T) {
	header := mail.Header{}

	err := SetSingleAddress(&header, "From", "John Doe <john@example.com")
	if err == nil {
		t.Fatalf("expected invalid address error")
	}
}
