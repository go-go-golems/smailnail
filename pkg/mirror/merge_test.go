package mirror

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestMergeServiceDryRunDiscoversShards(t *testing.T) {
	inputRoot := t.TempDir()
	createMergeTestShard(t, inputRoot, "2026-02", 21, []uint32{1, 2})
	createMergeTestShard(t, inputRoot, "2026-03", 21, []uint32{3})

	service := NewMergeService()
	report, err := service.Merge(t.Context(), MergeOptions{
		InputRoot:        inputRoot,
		OutputSQLitePath: filepath.Join(t.TempDir(), "merged.sqlite"),
		OutputMirrorRoot: filepath.Join(t.TempDir(), "merged-root"),
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if report.Status != "plan" {
		t.Fatalf("expected dry-run status plan, got %+v", report)
	}
	if report.ShardsDiscovered != 2 {
		t.Fatalf("expected 2 shards discovered, got %+v", report)
	}
	if report.MessagesScanned != 3 {
		t.Fatalf("expected 3 scanned messages, got %+v", report)
	}
	if len(report.Shards) != 2 {
		t.Fatalf("expected shard details, got %+v", report)
	}

	names := []string{report.Shards[0].Name, report.Shards[1].Name}
	if got := joinStrings(names); got != "2026-02,2026-03" {
		t.Fatalf("unexpected shard names %q", got)
	}
}

func TestMergeServiceDryRunFlagsMissingRawRootAsWarning(t *testing.T) {
	inputRoot := t.TempDir()
	shard := createMergeTestShard(t, inputRoot, "2026-03", 21, []uint32{1})
	if err := osRemoveAll(shard.RawRoot); err != nil {
		t.Fatalf("remove raw root error = %v", err)
	}

	service := NewMergeService()
	report, err := service.Merge(t.Context(), MergeOptions{
		InputRoot:        inputRoot,
		OutputSQLitePath: filepath.Join(t.TempDir(), "merged.sqlite"),
		OutputMirrorRoot: filepath.Join(t.TempDir(), "merged-root"),
		DryRun:           true,
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if len(report.Warnings) != 1 {
		t.Fatalf("expected one warning, got %+v", report)
	}
	if !report.Shards[0].MissingRawRoot {
		t.Fatalf("expected shard to flag missing raw root, got %+v", report.Shards[0])
	}
}

func TestMergeServiceRejectsUIDValidityConflicts(t *testing.T) {
	inputRoot := t.TempDir()
	createMergeTestShard(t, inputRoot, "2026-02", 21, []uint32{1})
	createMergeTestShard(t, inputRoot, "2026-03", 99, []uint32{2})

	service := NewMergeService()
	_, err := service.Merge(t.Context(), MergeOptions{
		InputRoot:        inputRoot,
		OutputSQLitePath: filepath.Join(t.TempDir(), "merged.sqlite"),
		OutputMirrorRoot: filepath.Join(t.TempDir(), "merged-root"),
		DryRun:           true,
	})
	if err == nil {
		t.Fatalf("expected UIDVALIDITY conflict")
	}
}

func TestDiscoverMergeShardsAppliesGlob(t *testing.T) {
	inputRoot := t.TempDir()
	createMergeTestShard(t, inputRoot, "2026-02", 21, []uint32{1})
	createMergeTestShard(t, inputRoot, "2026-03", 21, []uint32{2})

	shards, err := discoverMergeShards(inputRoot, "2026-03")
	if err != nil {
		t.Fatalf("discoverMergeShards() error = %v", err)
	}
	if len(shards) != 1 || shards[0].Name != "2026-03" {
		t.Fatalf("unexpected shards %+v", shards)
	}
}

func createMergeTestShard(t *testing.T, inputRoot, name string, uidValidity uint32, uids []uint32) MergeShardInfo {
	t.Helper()

	root := filepath.Join(inputRoot, name)
	store, err := OpenStore(filepath.Join(root, "mirror.sqlite"))
	if err != nil {
		t.Fatalf("OpenStore() error = %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	accountKey := AccountKey("localhost", 993, "merge-user")
	now := mustParseMergeTime(t, "2026-04-02T12:00:00Z")
	for _, uid := range uids {
		rawResult, err := WriteRawMessage(root, accountKey, "INBOX", uidValidity, uid, []byte("From: Tester <test@example.com>\r\nSubject: Merge\r\n\r\nBody\r\n"))
		if err != nil {
			t.Fatalf("WriteRawMessage() error = %v", err)
		}
		record := MessageRecord{
			AccountKey:   accountKey,
			MailboxName:  "INBOX",
			UIDValidity:  uidValidity,
			UID:          uid,
			MessageID:    "msg",
			InternalDate: now.Format(time.RFC3339),
			SentDate:     now.Format(time.RFC3339),
			Subject:      "Merge",
			FlagsJSON:    "[]",
			HeadersJSON:  "{}",
			PartsJSON:    "[]",
			BodyText:     "Body",
			SearchText:   "Merge\nBody",
			RawPath:      rawResult.Path,
			RawSHA256:    rawResult.SHA256,
			FirstSeenAt:  &now,
			LastSyncedAt: &now,
		}
		tx, err := store.db.BeginTxx(t.Context(), nil)
		if err != nil {
			t.Fatalf("BeginTxx() error = %v", err)
		}
		if err := upsertMessageRecord(t.Context(), tx, record); err != nil {
			t.Fatalf("upsertMessageRecord() error = %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("Commit() error = %v", err)
		}
	}

	return MergeShardInfo{
		Name:       name,
		Root:       root,
		SQLitePath: filepath.Join(root, "mirror.sqlite"),
		RawRoot:    filepath.Join(root, "raw"),
	}
}

func mustParseMergeTime(t *testing.T, raw string) time.Time {
	t.Helper()

	ts, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		t.Fatalf("parse time %q: %v", raw, err)
	}
	return ts
}

func joinStrings(values []string) string {
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return strings.Join(copied, ",")
}

func osRemoveAll(path string) error {
	return os.RemoveAll(path)
}
