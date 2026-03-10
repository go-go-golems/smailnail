package mailutil

import (
	"fmt"

	"github.com/emersion/go-message/mail"
)

// SetSingleAddress parses an RFC 5322 address string and writes it as a single-address header.
func SetSingleAddress(header *mail.Header, fieldName string, value string) error {
	if value == "" {
		return nil
	}

	address, err := mail.ParseAddress(value)
	if err != nil {
		return fmt.Errorf("invalid %s address %q: %w", fieldName, value, err)
	}

	header.SetAddressList(fieldName, []*mail.Address{address})
	return nil
}
