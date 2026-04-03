package enrich

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

func RunAll(ctx context.Context, sqlitePath string, opts Options) (AllReport, error) {
	start := time.Now()

	if sqlitePath == "" {
		return AllReport{}, errors.New("sqlite path is required")
	}

	db, err := sqlx.Open("sqlite3", sqlitePath)
	if err != nil {
		return AllReport{}, errors.Wrap(err, "open sqlite database for enrichment")
	}
	defer func() {
		_ = db.Close()
	}()

	sendersReport, err := (&SenderEnricher{}).Enrich(ctx, db, opts)
	if err != nil {
		return AllReport{}, err
	}
	threadsReport, err := (&ThreadEnricher{}).Enrich(ctx, db, opts)
	if err != nil {
		return AllReport{}, err
	}
	unsubscribeReport, err := (&UnsubscribeEnricher{}).Enrich(ctx, db, opts)
	if err != nil {
		return AllReport{}, err
	}

	return AllReport{
		Senders:     sendersReport,
		Threads:     threadsReport,
		Unsubscribe: unsubscribeReport,
		ElapsedMS:   time.Since(start).Milliseconds(),
	}, nil
}
