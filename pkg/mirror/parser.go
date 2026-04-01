package mirror

import (
	"bytes"
	"html"
	"io"
	stdmail "net/mail"
	"net/textproto"
	"regexp"
	"sort"
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
	Headers        map[string]string
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
	headers, err := parseRawHeaders(raw)
	if err != nil {
		return nil, err
	}

	reader, err := mail.CreateReader(bytes.NewReader(raw))
	if err != nil {
		return nil, errors.Wrap(err, "create mail reader")
	}

	ret := &ParsedMessage{
		Headers:     headers,
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
	ret.Headers = normalizeParsedHeaders(ret.Headers, ret)

	return ret, nil
}

func parseRawHeaders(raw []byte) (map[string]string, error) {
	message, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, errors.Wrap(err, "read raw message headers")
	}

	if len(message.Header) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(message.Header))
	for key := range message.Header {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ret := make(map[string]string, len(keys))
	for _, key := range keys {
		values := message.Header[key]
		if len(values) == 0 {
			continue
		}
		ret[textproto.CanonicalMIMEHeaderKey(key)] = strings.Join(values, "\n")
	}

	return ret, nil
}

func normalizeParsedHeaders(headers map[string]string, parsed *ParsedMessage) map[string]string {
	if parsed == nil {
		return headers
	}

	if headers == nil {
		headers = map[string]string{}
	}

	headers["Message-Id"] = normalizeMessageID(firstNonEmpty(parsed.MessageID, headers["Message-Id"]))
	if parsed.Subject != "" {
		headers["Subject"] = parsed.Subject
	}
	if parsed.SentDate != "" {
		headers["Date"] = parsed.SentDate
	}
	if parsed.FromSummary != "" {
		headers["From"] = parsed.FromSummary
	}
	if parsed.ToSummary != "" {
		headers["To"] = parsed.ToSummary
	}
	if parsed.CCSummary != "" {
		headers["Cc"] = parsed.CCSummary
	}

	return headers
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
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
	return normalizeMessageID(messageID)
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
		ret = append(ret, summarizeAddress(address))
	}
	return strings.Join(ret, ", ")
}

func summarizeAddress(address *stdmail.Address) string {
	if address == nil {
		return ""
	}

	name := normalizeWhitespace(address.Name)
	if name == "" {
		return address.Address
	}

	return name + " <" + address.Address + ">"
}

func normalizeMessageID(messageID string) string {
	messageID = strings.TrimSpace(messageID)
	messageID = strings.Trim(messageID, "<>")
	if messageID == "" {
		return ""
	}

	return "<" + messageID + ">"
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
