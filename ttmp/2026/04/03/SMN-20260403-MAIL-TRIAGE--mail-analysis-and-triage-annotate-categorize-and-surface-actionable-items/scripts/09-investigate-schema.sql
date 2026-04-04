-- 09-investigate-schema.sql
-- First exploration: understand the database schema
-- Run with: sqlite3 ~/smailnail/smailnail-last-24-months-merged.sqlite < thisfile.sql

.tables

.schema messages
.schema annotations
.schema senders
.schema threads
.schema annotation_logs
.schema target_groups
.schema target_group_members
.schema annotation_log_targets
