package annotate

import (
	"context"
	"path/filepath"
	"strings"

	annotatepkg "github.com/go-go-golems/smailnail/pkg/annotate"
	"github.com/go-go-golems/smailnail/pkg/mirror"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func openAnnotateRepo(ctx context.Context, sqlitePath string) (*annotatepkg.Repository, func(), error) {
	store, err := mirror.OpenStore(sqlitePath)
	if err != nil {
		return nil, nil, err
	}

	bootstrapRoot := filepath.Join(filepath.Dir(sqlitePath), ".smailnail-annotate")
	if _, err := store.Bootstrap(ctx, bootstrapRoot); err != nil {
		_ = store.Close()
		return nil, nil, err
	}
	if err := store.Close(); err != nil {
		return nil, nil, errors.Wrap(err, "close bootstrapped store handle")
	}

	db, err := sqlx.Open("sqlite3", sqlitePath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "open sqlite db for annotate command")
	}
	cleanup := func() {
		_ = db.Close()
	}
	return annotatepkg.NewRepository(db), cleanup, nil
}

func requireSQLitePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("sqlite-path is required")
	}
	return nil
}
