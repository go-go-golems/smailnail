package mirror

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func AccountKey(server string, port int, username string) string {
	source := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s-%d-%s", server, port, username)))
	return slugWithHash(source, "account")
}

func MailboxSlug(mailboxName string) string {
	return slugWithHash(strings.ToLower(strings.TrimSpace(mailboxName)), "mailbox")
}

func RawMessagePath(accountKey, mailboxName string, uidValidity, uid uint32) string {
	return filepath.Join(
		"raw",
		accountKey,
		MailboxSlug(mailboxName),
		fmt.Sprintf("%d", uidValidity),
		fmt.Sprintf("%d.eml", uid),
	)
}

func WriteRawMessage(mirrorRoot, accountKey, mailboxName string, uidValidity, uid uint32, raw []byte) (*RawMessageResult, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("raw message for uid %d is empty", uid)
	}

	relativePath := RawMessagePath(accountKey, mailboxName, uidValidity, uid)
	absolutePath := filepath.Join(filepath.Clean(mirrorRoot), relativePath)
	dir := filepath.Dir(absolutePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, errors.Wrap(err, "create raw message directory")
	}

	sum := sha256.Sum256(raw)
	sha := hex.EncodeToString(sum[:])

	if existing, err := os.ReadFile(absolutePath); err == nil {
		existingSum := sha256.Sum256(existing)
		if hex.EncodeToString(existingSum[:]) == sha {
			return &RawMessageResult{
				Path:         relativePath,
				SHA256:       sha,
				BytesWritten: len(raw),
				Reused:       true,
			}, nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, errors.Wrap(err, "read existing raw message")
	}

	tmpFile, err := os.CreateTemp(dir, "message-*.tmp")
	if err != nil {
		return nil, errors.Wrap(err, "create temporary raw message file")
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(raw); err != nil {
		_ = tmpFile.Close()
		return nil, errors.Wrap(err, "write temporary raw message file")
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return nil, errors.Wrap(err, "sync temporary raw message file")
	}
	if err := tmpFile.Close(); err != nil {
		return nil, errors.Wrap(err, "close temporary raw message file")
	}
	if err := os.Rename(tmpName, absolutePath); err != nil {
		return nil, errors.Wrap(err, "rename temporary raw message file")
	}

	return &RawMessageResult{
		Path:         relativePath,
		SHA256:       sha,
		BytesWritten: len(raw),
		Reused:       false,
	}, nil
}

func slugWithHash(source, fallback string) string {
	slug := sanitizePathComponent(source)
	if slug == "" {
		slug = fallback
	}

	sum := sha256.Sum256([]byte(source))
	hashPart := hex.EncodeToString(sum[:])[:12]

	const maxSlugLen = 48
	if len(slug) > maxSlugLen {
		slug = slug[:maxSlugLen]
		slug = strings.Trim(slug, "-")
	}

	return fmt.Sprintf("%s-%s", slug, hashPart)
}

func sanitizePathComponent(input string) string {
	var builder strings.Builder
	lastDash := false

	for _, r := range input {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}
