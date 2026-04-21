#!/usr/bin/env python3
"""Copy a mirror SQLite DB and seed pending review annotations for UI testing.

Default behavior matches the ad-hoc test data created during ticket follow-up work:
- copies the 2026-02 backfill mirror DB into repo-local tmp/
- inserts 20 pending sender annotations across 2 synthetic review runs
- ensures matching sender rows exist

Example:
    python scripts/seed_review_test_annotations.py

Custom source/dest:
    python scripts/seed_review_test_annotations.py \
      --source ~/smailnail/smailnail-last-24-months-backfill/2026-02/mirror.sqlite \
      --dest /tmp/local-review-test.sqlite
"""

from __future__ import annotations

import argparse
import shutil
import sqlite3
import uuid
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path


@dataclass(frozen=True)
class RunSeed:
    run_id: str
    source_label: str
    tag_prefix: str
    emails: list[str]


DEFAULT_RUNS = [
    RunSeed(
        run_id="test-run-review-2026-02-a",
        source_label="Backfill review smoke test A",
        tag_prefix="newsletter-review",
        emails=[
            "alpha@example.com",
            "bravo@example.com",
            "charlie@example.com",
            "delta@example.com",
            "echo@example.com",
            "foxtrot@example.com",
            "golf@example.com",
            "hotel@example.com",
            "india@example.com",
            "juliet@example.com",
        ],
    ),
    RunSeed(
        run_id="test-run-review-2026-02-b",
        source_label="Backfill review smoke test B",
        tag_prefix="promo-review",
        emails=[
            "kilo@example.com",
            "lima@example.com",
            "mike@example.com",
            "november@example.com",
            "oscar@example.com",
            "papa@example.com",
            "quebec@example.com",
            "romeo@example.com",
            "sierra@example.com",
            "tango@example.com",
        ],
    ),
]


def build_parser() -> argparse.ArgumentParser:
    repo_root = Path(__file__).resolve().parents[7]
    parser = argparse.ArgumentParser(
        description="Copy a mirror sqlite DB and seed pending review annotations.",
    )
    parser.add_argument(
        "--source",
        type=Path,
        default=Path.home()
        / "smailnail"
        / "smailnail-last-24-months-backfill"
        / "2026-02"
        / "mirror.sqlite",
        help="Source mirror.sqlite to copy before seeding.",
    )
    parser.add_argument(
        "--dest",
        type=Path,
        default=repo_root / "tmp" / "local-review-test-2026-02.sqlite",
        help="Destination SQLite file to create/update.",
    )
    parser.add_argument(
        "--no-copy",
        action="store_true",
        help="Skip copying source to dest and seed in-place at --dest.",
    )
    parser.add_argument(
        "--keep-existing-seeded-runs",
        action="store_true",
        help="Do not delete previous synthetic runs before inserting fresh rows.",
    )
    return parser


def copy_db(source: Path, dest: Path) -> None:
    dest.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(source, dest)


def seed(dest: Path, replace_existing: bool) -> tuple[int, int]:
    conn = sqlite3.connect(dest)
    try:
        cur = conn.cursor()
        now = datetime.now(UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")

        run_ids = [run.run_id for run in DEFAULT_RUNS]
        if replace_existing:
            placeholders = ", ".join("?" for _ in run_ids)
            cur.execute(
                f"DELETE FROM annotations WHERE agent_run_id IN ({placeholders})",
                run_ids,
            )
            cur.execute(
                f"DELETE FROM run_guideline_links WHERE agent_run_id IN ({placeholders})",
                run_ids,
            )

        sender_rows: list[tuple[str, str, str, int, str, str, str]] = []
        annotation_rows: list[tuple[str, str, str, str, str, str, str, str, str, str, str, str]] = []

        for run_index, run in enumerate(DEFAULT_RUNS, start=1):
            for item_index, email in enumerate(run.emails, start=1):
                sender_rows.append(
                    (
                        email,
                        f"Test Sender {run_index}-{item_index:02d}",
                        email.split("@", 1)[1],
                        1,
                        "2026-02-15",
                        "2026-02-15",
                        now,
                    )
                )
                annotation_rows.append(
                    (
                        str(uuid.uuid4()),
                        "sender",
                        email,
                        f"{run.tag_prefix}-{(item_index % 4) + 1}",
                        f"Synthetic pending review annotation {run_index}-{item_index:02d} for {email}",
                        "agent",
                        run.source_label,
                        run.run_id,
                        "to_review",
                        "seed-script",
                        now,
                        now,
                    )
                )

        cur.executemany(
            """
            INSERT INTO senders (
                email, display_name, domain, msg_count, first_seen_date, last_seen_date, last_synced_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(email) DO UPDATE SET
                display_name = excluded.display_name,
                domain = excluded.domain,
                msg_count = MAX(senders.msg_count, excluded.msg_count),
                last_seen_date = excluded.last_seen_date,
                last_synced_at = excluded.last_synced_at
            """,
            sender_rows,
        )

        cur.executemany(
            """
            INSERT INTO annotations (
                id, target_type, target_id, tag, note_markdown, source_kind, source_label,
                agent_run_id, review_state, created_by, created_at, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
            annotation_rows,
        )

        conn.commit()
        annotation_count = cur.execute(
            "SELECT COUNT(*) FROM annotations WHERE agent_run_id IN (?, ?)",
            run_ids,
        ).fetchone()[0]
        run_count = cur.execute(
            "SELECT COUNT(DISTINCT agent_run_id) FROM annotations WHERE agent_run_id IN (?, ?)",
            run_ids,
        ).fetchone()[0]
        return int(annotation_count), int(run_count)
    finally:
        conn.close()


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()

    source = args.source.expanduser().resolve()
    dest = args.dest.expanduser().resolve()

    if not args.no_copy:
        if not source.exists():
            parser.error(f"source DB does not exist: {source}")
        copy_db(source, dest)
    elif not dest.exists():
        parser.error(f"--no-copy was set but destination DB does not exist: {dest}")

    annotation_count, run_count = seed(
        dest,
        replace_existing=not args.keep_existing_seeded_runs,
    )

    print(f"seeded_annotations={annotation_count}")
    print(f"seeded_runs={run_count}")
    print(dest)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
