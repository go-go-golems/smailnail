package mirror

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-message/mail"

	"github.com/go-go-golems/smailnail/pkg/mailruntime"
	"github.com/go-go-golems/smailnail/pkg/mailutil"
)

func TestParseMessageMultipartAlternative(t *testing.T) {
	raw := mustCreateAlternativeMessage(t,
		"Parser Subject",
		"Plain parser body",
		"<html><body><p>Hello <strong>parser</strong></p></body></html>",
	)

	parsed, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}
	if parsed.Subject != "Parser Subject" {
		t.Fatalf("parsed.Subject = %q, want %q", parsed.Subject, "Parser Subject")
	}
	if parsed.BodyText != "Plain parser body" {
		t.Fatalf("parsed.BodyText = %q, want %q", parsed.BodyText, "Plain parser body")
	}
	if !strings.Contains(parsed.BodyHTML, "<strong>parser</strong>") {
		t.Fatalf("expected HTML body to be preserved, got %q", parsed.BodyHTML)
	}
	if !strings.Contains(parsed.SearchText, "Hello parser") {
		t.Fatalf("expected stripped HTML text in search text, got %q", parsed.SearchText)
	}
	if len(parsed.Parts) != 2 {
		t.Fatalf("expected two inline parts, got %d", len(parsed.Parts))
	}
}

func TestBuildMessageRecordUsesParsedProjection(t *testing.T) {
	raw := mustCreateAttachmentMessage(t,
		"Attachment Subject",
		"Attachment body text",
		"invoice.txt",
		[]byte("invoice payload"),
	)

	record, err := buildMessageRecord(
		"account-key",
		"INBOX",
		55,
		&mailruntime.FetchedMessage{
			UID:          7,
			Flags:        []string{"\\Seen"},
			Size:         int64(len(raw)),
			InternalDate: time.Date(2026, 4, 1, 21, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Envelope: &mailruntime.MessageEnvelope{
				Subject: "Fallback Subject",
			},
			Headers: map[string]string{"Subject": "Fallback Subject"},
			BodyRaw: raw,
		},
		&RawMessageResult{Path: "raw/account/inbox/55/7.eml", SHA256: "sha"},
		time.Date(2026, 4, 1, 21, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("buildMessageRecord() error = %v", err)
	}

	if record.Subject != "Attachment Subject" {
		t.Fatalf("record.Subject = %q, want parsed subject", record.Subject)
	}
	if record.BodyText != "Attachment body text" {
		t.Fatalf("record.BodyText = %q, want parsed plain text", record.BodyText)
	}
	if !record.HasAttachments {
		t.Fatalf("expected parsed attachment metadata to mark record as having attachments")
	}
	if !strings.Contains(record.PartsJSON, "invoice.txt") {
		t.Fatalf("expected parts JSON to include filename, got %s", record.PartsJSON)
	}
}

func mustCreateAlternativeMessage(t *testing.T, subject, textBody, htmlBody string) []byte {
	t.Helper()

	var buf bytes.Buffer
	header := mail.Header{}
	header.SetDate(time.Date(2026, 4, 1, 21, 0, 0, 0, time.UTC))
	if err := mailutil.SetSingleAddress(&header, "From", "Parser <parser@example.com>"); err != nil {
		t.Fatalf("SetSingleAddress From error = %v", err)
	}
	if err := mailutil.SetSingleAddress(&header, "To", "Reader <reader@example.com>"); err != nil {
		t.Fatalf("SetSingleAddress To error = %v", err)
	}
	header.SetSubject(subject)
	header.SetMessageID("<parser@example.com>")

	writer, err := mail.CreateWriter(&buf, header)
	if err != nil {
		t.Fatalf("CreateWriter() error = %v", err)
	}
	inlineWriter, err := writer.CreateInline()
	if err != nil {
		t.Fatalf("CreateInline() error = %v", err)
	}

	textHeader := mail.InlineHeader{}
	textHeader.Set("Content-Type", "text/plain; charset=utf-8")
	textWriter, err := inlineWriter.CreatePart(textHeader)
	if err != nil {
		t.Fatalf("CreatePart(text) error = %v", err)
	}
	if _, err := textWriter.Write([]byte(textBody)); err != nil {
		t.Fatalf("Write(text) error = %v", err)
	}
	if err := textWriter.Close(); err != nil {
		t.Fatalf("Close(text) error = %v", err)
	}

	htmlHeader := mail.InlineHeader{}
	htmlHeader.Set("Content-Type", "text/html; charset=utf-8")
	htmlWriter, err := inlineWriter.CreatePart(htmlHeader)
	if err != nil {
		t.Fatalf("CreatePart(html) error = %v", err)
	}
	if _, err := htmlWriter.Write([]byte(htmlBody)); err != nil {
		t.Fatalf("Write(html) error = %v", err)
	}
	if err := htmlWriter.Close(); err != nil {
		t.Fatalf("Close(html) error = %v", err)
	}

	if err := inlineWriter.Close(); err != nil {
		t.Fatalf("Close(inlineWriter) error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close(writer) error = %v", err)
	}

	return buf.Bytes()
}

func mustCreateAttachmentMessage(t *testing.T, subject, body, filename string, attachment []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	header := mail.Header{}
	header.SetDate(time.Date(2026, 4, 1, 21, 0, 0, 0, time.UTC))
	if err := mailutil.SetSingleAddress(&header, "From", "Parser <parser@example.com>"); err != nil {
		t.Fatalf("SetSingleAddress From error = %v", err)
	}
	if err := mailutil.SetSingleAddress(&header, "To", "Reader <reader@example.com>"); err != nil {
		t.Fatalf("SetSingleAddress To error = %v", err)
	}
	header.SetSubject(subject)
	header.SetMessageID("<attachment@example.com>")

	writer, err := mail.CreateWriter(&buf, header)
	if err != nil {
		t.Fatalf("CreateWriter() error = %v", err)
	}
	inlineHeader := mail.InlineHeader{}
	inlineHeader.Set("Content-Type", "text/plain; charset=utf-8")
	inlineWriter, err := writer.CreateSingleInline(inlineHeader)
	if err != nil {
		t.Fatalf("CreateSingleInline() error = %v", err)
	}
	if _, err := inlineWriter.Write([]byte(body)); err != nil {
		t.Fatalf("Write(body) error = %v", err)
	}
	if err := inlineWriter.Close(); err != nil {
		t.Fatalf("Close(body) error = %v", err)
	}

	attachmentHeader := mail.AttachmentHeader{}
	attachmentHeader.Set("Content-Type", "text/plain")
	attachmentHeader.SetFilename(filename)
	attachmentWriter, err := writer.CreateAttachment(attachmentHeader)
	if err != nil {
		t.Fatalf("CreateAttachment() error = %v", err)
	}
	if _, err := attachmentWriter.Write(attachment); err != nil {
		t.Fatalf("Write(attachment) error = %v", err)
	}
	if err := attachmentWriter.Close(); err != nil {
		t.Fatalf("Close(attachment) error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close(writer) error = %v", err)
	}

	return buf.Bytes()
}
