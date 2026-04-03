package sqlite

import (
	"context"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/smailnail/pkg/annotationui"
	"github.com/go-go-golems/smailnail/pkg/mirror"
	"github.com/go-go-golems/smailnail/pkg/smailnaild/web"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ServeCommand struct {
	*cmds.CommandDescription
}

type serveSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	MirrorRoot string `glazed:"mirror-root"`
	ListenHost string `glazed:"listen-host"`
	ListenPort int    `glazed:"listen-port"`
	QueryDir   string `glazed:"query-dir"`
	PresetDir  string `glazed:"preset-dir"`
}

var _ cmds.BareCommand = &ServeCommand{}

func NewServeCommand() (*ServeCommand, error) {
	section, err := schema.NewSection(
		schema.DefaultSlug,
		"SQLite Serve Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault(mirror.DefaultSQLiteDBPath)),
			fields.New("mirror-root", fields.TypeString, fields.WithHelp("Mirror root used when bootstrapping the sqlite store (defaults next to the sqlite file)")),
			fields.New("listen-host", fields.TypeString, fields.WithHelp("Host interface to bind"), fields.WithDefault("0.0.0.0")),
			fields.New("listen-port", fields.TypeInteger, fields.WithHelp("Port to listen on"), fields.WithDefault(8080)),
			fields.New("query-dir", fields.TypeString, fields.WithHelp("Directory for saved SQL queries (defaults next to the sqlite file)")),
			fields.New("preset-dir", fields.TypeString, fields.WithHelp("Optional directory containing additional preset .sql files")),
		),
	)
	if err != nil {
		return nil, err
	}

	return &ServeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"serve",
			cmds.WithShort("Start the sqlite-backed annotation UI backend"),
			cmds.WithLong(`Start the sqlite-backed annotation UI backend.

This server is separate from smailnaild. It serves the annotation review UI,
sender browser, agent run views, and read-only SQL workbench directly from a
mirror sqlite database.`),
			cmds.WithSections(section),
		),
	}, nil
}

func (c *ServeCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	settings := &serveSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}
	if settings.SQLitePath == "" {
		return errors.New("sqlite-path is required")
	}

	mirrorRoot := settings.MirrorRoot
	if mirrorRoot == "" {
		mirrorRoot = filepath.Join(filepath.Dir(settings.SQLitePath), ".smailnail-sqlite")
	}
	queryDir := settings.QueryDir
	if queryDir == "" {
		queryDir = filepath.Join(filepath.Dir(settings.SQLitePath), ".smailnail-queries")
	}

	store, err := mirror.OpenStore(settings.SQLitePath)
	if err != nil {
		return errors.Wrap(err, "open mirror store")
	}
	if _, err := store.Bootstrap(ctx, mirrorRoot); err != nil {
		_ = store.Close()
		return errors.Wrap(err, "bootstrap mirror store")
	}
	if err := store.Close(); err != nil {
		return errors.Wrap(err, "close mirror store bootstrap handle")
	}

	db, err := sqlx.Open("sqlite3", settings.SQLitePath)
	if err != nil {
		return errors.Wrap(err, "open sqlite database")
	}
	defer func() { _ = db.Close() }()

	server := annotationui.NewHTTPServer(annotationui.ServerOptions{
		Host: settings.ListenHost,
		Port: settings.ListenPort,
		DB:   db,
		DBInfo: annotationui.DatabaseInfo{
			Driver: "sqlite3",
			Target: settings.SQLitePath,
			Mode:   "mirror",
		},
		QueryDirs:  []string{queryDir},
		PresetDirs: []string{settings.PresetDir},
		PublicFS:   web.PublicFS,
	})

	return annotationui.RunServer(ctx, server)
}
