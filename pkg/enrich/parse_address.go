package enrich

import (
	"fmt"
	stdmail "net/mail"
	"strings"
	"unicode"

	gomail "github.com/emersion/go-message/mail"
)

const applePrivateRelayDomain = "privaterelay.appleid.com"

func ParseFromSummary(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("empty from_summary")
	}

	if addrs, err := gomail.ParseAddressList(raw); err == nil && len(addrs) > 0 {
		return strings.ToLower(strings.TrimSpace(addrs[0].Address)), strings.TrimSpace(addrs[0].Name), nil
	}

	if addr, err := gomail.ParseAddress(raw); err == nil && addr != nil {
		return strings.ToLower(strings.TrimSpace(addr.Address)), strings.TrimSpace(addr.Name), nil
	}

	if addr, err := stdmail.ParseAddress(raw); err == nil && addr != nil {
		return strings.ToLower(strings.TrimSpace(addr.Address)), strings.TrimSpace(addr.Name), nil
	}

	if strings.Contains(raw, "@") && !strings.ContainsAny(raw, "<>,") {
		return strings.ToLower(raw), "", nil
	}

	return "", "", fmt.Errorf("parse from_summary %q", raw)
}

func ExtractDomain(email string) string {
	_, domain, ok := strings.Cut(strings.ToLower(strings.TrimSpace(email)), "@")
	if !ok {
		return ""
	}
	return domain
}

func IsPrivateRelay(domain string) bool {
	return strings.EqualFold(strings.TrimSpace(domain), applePrivateRelayDomain)
}

func GuessRelayDomain(displayName string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(displayName)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
