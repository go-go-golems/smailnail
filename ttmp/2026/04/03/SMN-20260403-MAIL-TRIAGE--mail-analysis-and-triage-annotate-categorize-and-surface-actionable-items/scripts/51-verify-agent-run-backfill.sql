-- 51-verify-agent-run-backfill.sql
-- Verify agent_run_id was backfilled and log-target links exist

-- Agent run IDs
SELECT agent_run_id, COUNT(*) as cnt FROM annotations GROUP BY agent_run_id;
SELECT agent_run_id, COUNT(*) as cnt FROM annotation_logs GROUP BY agent_run_id;

-- Annotation source metadata
SELECT source_kind, source_label, created_by, agent_run_id, COUNT(*) as cnt
FROM annotations
GROUP BY source_kind, source_label, created_by, agent_run_id;

-- Log entries
SELECT id, title, source_kind, source_label, created_by, agent_run_id, created_at
FROM annotation_logs ORDER BY created_at;

-- Log-target links
SELECT l.title, lt.target_type, lt.target_id
FROM annotation_log_targets lt
JOIN annotation_logs l ON l.id = lt.log_id
ORDER BY l.created_at, lt.target_type, lt.target_id;

-- Senders with multiple annotations
SELECT target_id, COUNT(*) as num_tags, GROUP_CONCAT(tag, ', ') as tags
FROM annotations
WHERE target_type = 'sender'
GROUP BY target_id
HAVING num_tags > 1
ORDER BY num_tags DESC;
