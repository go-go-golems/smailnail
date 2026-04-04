-- 52-clean-annotations.sql
-- Remove all annotations, logs, groups, and group members from a copy of the DB.
-- Run against the CLEAN copy only, never the original.
-- Usage: sqlite3 ~/smailnail/smailnail-last-24-months-clean.sqlite < this_file.sql

DELETE FROM annotation_log_targets;
DELETE FROM annotation_logs;
DELETE FROM target_group_members;
DELETE FROM target_groups;
DELETE FROM annotations;

-- Verify
SELECT 'annotations' as tbl, COUNT(*) as cnt FROM annotations
UNION ALL SELECT 'annotation_logs', COUNT(*) FROM annotation_logs
UNION ALL SELECT 'annotation_log_targets', COUNT(*) FROM annotation_log_targets
UNION ALL SELECT 'target_groups', COUNT(*) FROM target_groups
UNION ALL SELECT 'target_group_members', COUNT(*) FROM target_group_members;
