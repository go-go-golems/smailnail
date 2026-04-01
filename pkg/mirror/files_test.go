package mirror

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteRawMessageIsIdempotent(t *testing.T) {
	root := t.TempDir()
	accountKey := AccountKey("imap.example.com", 993, "user@example.com")

	first, err := WriteRawMessage(root, accountKey, "INBOX/Receipts", 42, 7, []byte("hello world"))
	if err != nil {
		t.Fatalf("WriteRawMessage() error = %v", err)
	}
	if first.Reused {
		t.Fatalf("expected first write to create a new file")
	}

	second, err := WriteRawMessage(root, accountKey, "INBOX/Receipts", 42, 7, []byte("hello world"))
	if err != nil {
		t.Fatalf("WriteRawMessage() second error = %v", err)
	}
	if !second.Reused {
		t.Fatalf("expected second write to reuse the existing file")
	}
	if first.Path != second.Path {
		t.Fatalf("expected same relative path, got %q and %q", first.Path, second.Path)
	}

	rawPath := filepath.Join(root, second.Path)
	raw, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(raw) != "hello world" {
		t.Fatalf("unexpected raw file contents: %q", string(raw))
	}
	if !strings.Contains(second.Path, MailboxSlug("INBOX/Receipts")) {
		t.Fatalf("expected mailbox slug in path %q", second.Path)
	}
}

func TestAccountKeyAndMailboxSlugAreStableASCII(t *testing.T) {
	accountKey := AccountKey("imap.EXAMPLE.com", 993, "User Name")
	mailboxSlug := MailboxSlug("INBOX/Reports & Billing")

	for _, value := range []string{accountKey, mailboxSlug} {
		if value == "" {
			t.Fatalf("expected non-empty slug")
		}
		if strings.ContainsAny(value, " /&") {
			t.Fatalf("expected sanitized slug, got %q", value)
		}
	}
}
