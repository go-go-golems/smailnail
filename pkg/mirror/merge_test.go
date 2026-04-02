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

func TestMergeServiceMergesShardsAndRebuildsDerivedState(t *testing.T) {
	inputRoot := t.TempDir()
	shardA := createMergeTestShard(t, inputRoot, "2026-02", 21, []uint32{1, 2})
	shardB := createMergeTestShard(t, inputRoot, "2026-03", 21, []uint32{3})

	outputDir := t.TempDir()
	outputSQLite := filepath.Join(outputDir, "merged.sqlite")
	outputRoot := filepath.Join(outputDir, "merged-root")

	service := NewMergeService()
	report, err := service.Merge(t.Context(), MergeOptions{
		InputRoot:        inputRoot,
		OutputSQLitePath: outputSQLite,
		OutputMirrorRoot: outputRoot,
		CopyRaw:          true,
		RebuildFTS:       true,
		RebuildSyncState: true,
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if report.Status != "merged" {
		t.Fatalf("expected merged status, got %+v", report)
	}
	if report.ShardsMerged != 2 {
		t.Fatalf("expected 2 merged shards, got %+v", report)
	}
	if report.MessagesInserted != 3 || report.MessagesUpdated != 0 {
		t.Fatalf("unexpected message counters %+v", report)
	}
	if report.RawFilesCopied != 3 || report.RawFilesReused != 0 {
		t.Fatalf("unexpected raw counters %+v", report)
	}
	if report.MailboxStatesRebuilt != 1 {
		t.Fatalf("expected one rebuilt mailbox state, got %+v", report)
	}
	if report.FTSRowsRebuilt != 3 {
		t.Fatalf("expected 3 rebuilt fts rows, got %+v", report)
	}

	store, err := OpenStore(outputSQLite)
	if err != nil {
		t.Fatalf("OpenStore() merged error = %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	if got := countMessages(t, store.db, "INBOX"); got != 3 {
		t.Fatalf("expected 3 merged messages, got %d", got)
	}
	state, err := loadMailboxSyncState(t.Context(), store.db, AccountKey("localhost", 993, "merge-user"), "INBOX")
	if err != nil {
		t.Fatalf("loadMailboxSyncState() error = %v", err)
	}
	if state == nil || state.HighestUID != 3 || state.LastUIDNext != 4 {
		t.Fatalf("unexpected rebuilt mailbox state %+v", state)
	}

	assertFileExists(t, filepath.Join(outputRoot, RawMessagePath(AccountKey("localhost", 993, "merge-user"), "INBOX", 21, 1)))
	assertFileExists(t, filepath.Join(outputRoot, RawMessagePath(AccountKey("localhost", 993, "merge-user"), "INBOX", 21, 3)))

	assertFileExists(t, filepath.Join(shardA.RawRoot, RawMessagePath(AccountKey("localhost", 993, "merge-user"), "INBOX", 21, 1)))
	assertFileExists(t, filepath.Join(shardB.RawRoot, RawMessagePath(AccountKey("localhost", 993, "merge-user"), "INBOX", 21, 3)))

	var ftsCount int
	if err := store.db.Get(&ftsCount, `SELECT COUNT(*) FROM messages_fts`); err != nil {
		t.Fatalf("count messages_fts error = %v", err)
	}
	if ftsCount != 3 {
		t.Fatalf("expected 3 messages_fts rows, got %d", ftsCount)
	}
}

func TestMergeServiceWarnsOnMissingRawByDefault(t *testing.T) {
	inputRoot := t.TempDir()
	shard := createMergeTestShard(t, inputRoot, "2026-03", 21, []uint32{1})
	missingPath := filepath.Join(shard.RawRoot, RawMessagePath(AccountKey("localhost", 993, "merge-user"), "INBOX", 21, 1))
	if err := os.Remove(missingPath); err != nil {
		t.Fatalf("remove raw file error = %v", err)
	}

	outputDir := t.TempDir()
	service := NewMergeService()
	report, err := service.Merge(t.Context(), MergeOptions{
		InputRoot:        inputRoot,
		OutputSQLitePath: filepath.Join(outputDir, "merged.sqlite"),
		OutputMirrorRoot: filepath.Join(outputDir, "merged-root"),
		CopyRaw:          true,
	})
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if report.RawFilesMissing != 1 {
		t.Fatalf("expected one missing raw file, got %+v", report)
	}
	if len(report.Warnings) == 0 {
		t.Fatalf("expected missing raw warning, got %+v", report)
	}
}

func TestMergeServiceFailsWhenMissingRawIsStrict(t *testing.T) {
	inputRoot := t.TempDir()
	shard := createMergeTestShard(t, inputRoot, "2026-03", 21, []uint32{1})
	missingPath := filepath.Join(shard.RawRoot, RawMessagePath(AccountKey("localhost", 993, "merge-user"), "INBOX", 21, 1))
	if err := os.Remove(missingPath); err != nil {
		t.Fatalf("remove raw file error = %v", err)
	}

	outputDir := t.TempDir()
	service := NewMergeService()
	_, err := service.Merge(t.Context(), MergeOptions{
		InputRoot:        inputRoot,
		OutputSQLitePath: filepath.Join(outputDir, "merged.sqlite"),
		OutputMirrorRoot: filepath.Join(outputDir, "merged-root"),
		CopyRaw:          true,
		FailOnMissingRaw: true,
		RebuildFTS:       true,
		RebuildSyncState: true,
	})
	if err == nil {
		t.Fatalf("expected strict missing raw error")
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
		rawResult, err := WriteRawMessage(filepath.Join(root, "raw"), accountKey, "INBOX", uidValidity, uid, []byte("From: Tester <test@example.com>\r\nSubject: Merge\r\n\r\nBody\r\n"))
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
