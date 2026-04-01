package mirror

import (
	"bytes"
	"html"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/pkg/errors"
)

type ParsedPart struct {
	ContentType string `json:"contentType"`
	Disposition string `json:"disposition"`
	Filename    string `json:"filename,omitempty"`
	Charset     string `json:"charset,omitempty"`
	SizeBytes   int    `json:"sizeBytes"`
}

type ParsedMessage struct {
	Subject        string
	MessageID      string
	SentDate       string
	FromSummary    string
	ToSummary      string
	CCSummary      string
	BodyText       string
	BodyHTML       string
	SearchText     string
	HasAttachments bool
	Parts          []ParsedPart
}

func ParseMessage(raw []byte) (*ParsedMessage, error) {
	reader, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		return nil, errors.Wrap(err, "create mail reader")
	}

	ret := &ParsedMessage{
		Subject:     readHeaderSubject(reader.Header),
		MessageID:   readHeaderMessageID(reader.Header),
		SentDate:    readHeaderDate(reader.Header),
		FromSummary: readHeaderAddressList(reader.Header, "From"),
		ToSummary:   readHeaderAddressList(reader.Header, "To"),
		CCSummary:   readHeaderAddressList(reader.Header, "Cc"),
	}

	var plainParts []string
	var htmlParts []string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "read message part")
		}

		body, err := io.ReadAll(part.Body)
		if err != nil {
			return nil, errors.Wrap(err, "read message part body")
		}

		switch header := part.Header.(type) {
		case *mail.InlineHeader:
			contentType, params, _ := header.ContentType()
			ret.Parts = append(ret.Parts, ParsedPart{
				ContentType: contentType,
				Disposition: "inline",
				Charset:     params["charset"],
				SizeBytes:   len(body),
			})
			switch {
			case strings.HasPrefix(contentType, "text/plain"):
				plainParts = append(plainParts, strings.TrimSpace(string(body)))
			case strings.HasPrefix(contentType, "text/html"):
				htmlParts = append(htmlParts, strings.TrimSpace(string(body)))
			}
		case *mail.AttachmentHeader:
			contentType, params, _ := header.ContentType()
			filename, _ := header.Filename()
			ret.Parts = append(ret.Parts, ParsedPart{
				ContentType: contentType,
				Disposition: "attachment",
				Filename:    filename,
				Charset:     params["charset"],
				SizeBytes:   len(body),
			})
			ret.HasAttachments = true
		}
	}

	ret.BodyText = joinNonEmpty(plainParts)
	ret.BodyHTML = joinNonEmpty(htmlParts)
	ret.SearchText = normalizeSearchText([]string{
		ret.Subject,
		ret.FromSummary,
		ret.ToSummary,
		ret.CCSummary,
		ret.BodyText,
		stripHTML(ret.BodyHTML),
	})

	return ret, nil
}

func readHeaderSubject(header mail.Header) string {
	subject, err := header.Subject()
	if err != nil {
		return ""
	}
	return subject
}

func readHeaderMessageID(header mail.Header) string {
	messageID, err := header.MessageID()
	if err != nil {
		return ""
	}
	return messageID
}

func readHeaderDate(header mail.Header) string {
	date, err := header.Date()
	if err != nil {
		return ""
	}
	return date.Format(time.RFC3339)
}

func readHeaderAddressList(header mail.Header, key string) string {
	addresses, err := header.AddressList(key)
	if err != nil || len(addresses) == 0 {
		return ""
	}

	ret := make([]string, 0, len(addresses))
	for _, address := range addresses {
		ret = append(ret, address.String())
	}
	return strings.Join(ret, ", ")
}

func joinNonEmpty(parts []string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, "\n\n")
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func stripHTML(input string) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}
	withoutTags := htmlTagRe.ReplaceAllString(input, " ")
	return normalizeWhitespace(html.UnescapeString(withoutTags))
}

func normalizeSearchText(parts []string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = normalizeWhitespace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, "\n")
}

func normalizeWhitespace(input string) string {
	if input == "" {
		return ""
	}
	return strings.Join(strings.Fields(input), " ")
}
