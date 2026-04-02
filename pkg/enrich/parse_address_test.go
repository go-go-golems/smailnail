package enrich

import "testing"

func TestParseFromSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		raw         string
		wantEmail   string
		wantDisplay string
		wantErr     bool
	}{
		{
			name:        "display and address",
			raw:         "Zillow <updates@example.com>",
			wantEmail:   "updates@example.com",
			wantDisplay: "Zillow",
		},
		{
			name:      "bare address",
			raw:       "sender@example.com",
			wantEmail: "sender@example.com",
		},
		{
			name:        "encoded display name",
			raw:         "=?utf-8?Q?The=20Providence=20Athen=C3=A6um?= <news@example.org>",
			wantEmail:   "news@example.org",
			wantDisplay: "The Providence Athenæum",
		},
		{
			name:        "private relay sender",
			raw:         "Zillow <instant-updates_at_example.com_foo@privaterelay.appleid.com>",
			wantEmail:   "instant-updates_at_example.com_foo@privaterelay.appleid.com",
			wantDisplay: "Zillow",
		},
		{
			name:    "invalid raw value",
			raw:     "not actually an address",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotEmail, gotDisplay, err := ParseFromSummary(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFromSummary(%q) expected error", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFromSummary(%q) error = %v", tt.raw, err)
			}
			if gotEmail != tt.wantEmail {
				t.Fatalf("ParseFromSummary(%q) email = %q, want %q", tt.raw, gotEmail, tt.wantEmail)
			}
			if gotDisplay != tt.wantDisplay {
				t.Fatalf("ParseFromSummary(%q) display = %q, want %q", tt.raw, gotDisplay, tt.wantDisplay)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	t.Parallel()

	if got := ExtractDomain("USER@Example.COM"); got != "example.com" {
		t.Fatalf("ExtractDomain() = %q, want %q", got, "example.com")
	}
	if got := ExtractDomain("not-an-email"); got != "" {
		t.Fatalf("ExtractDomain() = %q, want empty string", got)
	}
}

func TestPrivateRelayHelpers(t *testing.T) {
	t.Parallel()

	if !IsPrivateRelay("privaterelay.appleid.com") {
		t.Fatal("expected private relay domain to be detected")
	}
	if got := GuessRelayDomain("The Providence Athenæum"); got != "theprovidenceathenæum" {
		t.Fatalf("GuessRelayDomain() = %q", got)
	}
}
