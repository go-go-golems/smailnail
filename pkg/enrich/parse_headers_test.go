package enrich

import "testing"

func TestGetHeader(t *testing.T) {
	t.Parallel()

	headersJSON := `{"List-Unsubscribe":"<mailto:test@example.com>","Message-Id":"<abc@example.com>"}`

	if got := GetHeader(headersJSON, "list-unsubscribe"); got != "<mailto:test@example.com>" {
		t.Fatalf("GetHeader() = %q", got)
	}
	if got := GetHeader(headersJSON, "missing"); got != "" {
		t.Fatalf("GetHeader() = %q, want empty string", got)
	}
}

func TestParseListUnsubscribe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		raw          string
		post         string
		wantMailto   string
		wantHTTP     string
		wantOneClick bool
	}{
		{
			name:       "mailto only",
			raw:        "<mailto:unsubscribe@example.com?subject=unsubscribe>",
			wantMailto: "mailto:unsubscribe@example.com?subject=unsubscribe",
		},
		{
			name:     "http only",
			raw:      "<https://example.com/unsub?t=abc>",
			wantHTTP: "https://example.com/unsub?t=abc",
		},
		{
			name:         "both links and one click",
			raw:          "<mailto:unsubscribe@example.com>, <https://example.com/unsub?t=abc>",
			post:         "List-Unsubscribe=One-Click",
			wantMailto:   "mailto:unsubscribe@example.com",
			wantHTTP:     "https://example.com/unsub?t=abc",
			wantOneClick: true,
		},
		{
			name: "malformed input",
			raw:  "mailto:test@example.com, https://example.com/unsub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMailto, gotHTTP, gotOneClick := ParseListUnsubscribe(tt.raw, tt.post)
			if gotMailto != tt.wantMailto {
				t.Fatalf("mailto = %q, want %q", gotMailto, tt.wantMailto)
			}
			if gotHTTP != tt.wantHTTP {
				t.Fatalf("http = %q, want %q", gotHTTP, tt.wantHTTP)
			}
			if gotOneClick != tt.wantOneClick {
				t.Fatalf("oneClick = %t, want %t", gotOneClick, tt.wantOneClick)
			}
		})
	}
}
