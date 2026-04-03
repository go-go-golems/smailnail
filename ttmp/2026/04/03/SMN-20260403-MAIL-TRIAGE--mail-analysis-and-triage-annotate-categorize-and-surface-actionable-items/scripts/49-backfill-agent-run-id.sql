-- 49-backfill-agent-run-id.sql
-- Backfill agent_run_id on all annotations and logs from this session.
-- All rows with source_label='mail-triage-v1' belong to this run.
-- Using a fixed run ID: "mail-triage-v1-20260403"

UPDATE annotations
SET agent_run_id = 'mail-triage-v1-20260403'
WHERE source_label = 'mail-triage-v1' AND agent_run_id = '';

UPDATE annotation_logs
SET agent_run_id = 'mail-triage-v1-20260403'
WHERE source_label = 'mail-triage-v1' AND agent_run_id = '';
