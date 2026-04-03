-- 10-investigate-basic-stats.sql
-- Basic mailbox stats: totals, date range, mailboxes, top domains, enrichment status

-- Total messages
SELECT 'Total messages' as metric, COUNT(*) as value FROM messages;

-- Date range
SELECT 'Date range' as metric, MIN(internal_date) || ' to ' || MAX(internal_date) as value FROM messages;

-- Mailboxes
SELECT mailbox_name, COUNT(*) as cnt FROM messages GROUP BY mailbox_name ORDER BY cnt DESC LIMIT 30;

-- Top 20 sender domains
SELECT sender_domain, COUNT(*) as cnt FROM messages GROUP BY sender_domain ORDER BY cnt DESC LIMIT 20;

-- Enrichment status
SELECT 'Threads' as metric, COUNT(*) as value FROM threads;
SELECT 'Senders' as metric, COUNT(*) as value FROM senders;
SELECT 'Annotations' as metric, COUNT(*) as value FROM annotations;
SELECT 'Annotation logs' as metric, COUNT(*) as value FROM annotation_logs;
SELECT 'Groups' as metric, COUNT(*) as value FROM target_groups;
