-- 39-investigate-slow-query-analysis.sql
-- Performance analysis of the slow category-count query
-- 
-- ROOT CAUSE: No index on messages.sender_email
-- The join `annotations a ON a.target_id = m.sender_email` forces a full table SCAN
-- on messages (32,912 rows) for each of the ~446 annotations with target_type='sender'.
-- 
-- The EXPLAIN QUERY PLAN shows:
--   SEARCH a USING INDEX idx_annotations_target (target_type=?)
--   SCAN m                    <-- THIS IS THE PROBLEM
--   USE TEMP B-TREE FOR GROUP BY
--   USE TEMP B-TREE FOR count(DISTINCT)
--   USE TEMP B-TREE FOR ORDER BY
-- 
-- SQLite picks annotations as the outer table (correct, it's smaller: 446 rows),
-- but then for the join condition `a.target_id = m.sender_email` it has to scan
-- all 32,912 messages because there is NO INDEX on messages.sender_email.
-- 
-- Actually the real issue is worse: SQLite's query plan shows it scans m as outer
-- and searches a as inner. So it's doing 32,912 index lookups into annotations.
-- That's fast for the annotation side but the full scan of messages (which includes
-- reading body_text columns even though they're not needed) is I/O bound.
-- The messages table is large on disk because body_text and body_html are stored
-- inline — even a sequential scan touches all that data.
--
-- FIX (would require DB modification):
--   CREATE INDEX idx_messages_sender_email ON messages(sender_email);
-- 
-- This would let SQLite do:
--   1. Scan annotations WHERE target_type='sender' (446 rows via index)
--   2. For each, seek messages by sender_email index (fast lookup)
-- 
-- ALTERNATIVE APPROACH (no DB modification):
-- Pre-aggregate sender -> message count, then join to annotations:

-- Fast version using subquery (still ~5min without index, but conceptually cleaner):
EXPLAIN QUERY PLAN
WITH sender_counts AS (
  SELECT sender_email, COUNT(*) as msg_count
  FROM messages
  WHERE sender_email != ''
  GROUP BY sender_email
)
SELECT 
  CASE 
    WHEN a.tag LIKE 'important/%' THEN 'IMPORTANT'
    WHEN a.tag = 'personal' THEN 'PERSONAL'
    WHEN a.tag = 'work' THEN 'WORK'
    WHEN a.tag = 'community' THEN 'COMMUNITY'
    WHEN a.tag LIKE 'newsletter/%' THEN 'NEWSLETTER'
    WHEN a.tag LIKE 'noise/%' THEN 'NOISE'
    ELSE 'OTHER'
  END as category,
  SUM(sc.msg_count) as msgs
FROM annotations a
JOIN sender_counts sc ON sc.sender_email = a.target_id
WHERE a.target_type = 'sender'
GROUP BY category
ORDER BY msgs DESC;

-- Timing comparisons were done in bash:
-- Original query:          4m38s (SCAN m, COUNT DISTINCT)
-- CTE drive from annot:    4m33s (same plan, no improvement)
-- CTE no COUNT DISTINCT:   5m10s (actually slower, more temp B-trees)
--
-- CONCLUSION: All variants are ~5 minutes because the bottleneck is the
-- full table scan of messages. The only real fix is adding an index on
-- messages.sender_email. The table is ~2GB+ on disk due to body_text/html
-- columns, and every scan reads all of it.
