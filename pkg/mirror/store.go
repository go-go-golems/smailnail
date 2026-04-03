package mirror

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Store struct {
	db   *sqlx.DB
	path string
}

func OpenStore(path string) (*Store, error) {
	if path == "" {
		path = DefaultSQLiteDBPath
	}
	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		return nil, errors.Wrap(err, "open mirror sqlite db")
	}
	return &Store{
		db:   db,
		path: path,
	}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Bootstrap(ctx context.Context, mirrorRoot string) (*BootstrapReport, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("store is not open")
	}
	if mirrorRoot == "" {
		mirrorRoot = DefaultMirrorRoot
	}
	if err := EnsureMirrorRoot(mirrorRoot); err != nil {
		return nil, err
	}
	report, err := bootstrapSchema(ctx, s.db)
	if err != nil {
		return nil, err
	}
	report.Database = DatabaseInfo{
		Driver: "sqlite3",
		Path:   s.path,
	}
	report.MirrorRoot = mirrorRoot
	return &report, nil
}

func EnsureMirrorRoot(root string) error {
	if root == "" {
		root = DefaultMirrorRoot
	}
	cleanRoot := filepath.Clean(root)
	if err := os.MkdirAll(filepath.Join(cleanRoot, "raw"), 0o755); err != nil {
		return errors.Wrap(err, "create mirror root")
	}
	return nil
}
