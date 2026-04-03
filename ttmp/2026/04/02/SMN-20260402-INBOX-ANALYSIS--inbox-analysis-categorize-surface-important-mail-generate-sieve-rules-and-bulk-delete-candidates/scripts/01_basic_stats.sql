-- 01_basic_stats.sql
-- Basic statistics: counts, date ranges, size distribution, flags

-- Total messages
SELECT 'total_messages' AS metric, COUNT(*) AS value FROM messages;

-- Date range
SELECT 'earliest_sent' AS metric, MIN(sent_date) AS value FROM messages;
SELECT 'latest_sent'   AS metric, MAX(sent_date) AS value FROM messages;

-- Messages by month
SELECT
  substr(sent_date, 1, 7) AS month,
  COUNT(*) AS msg_count,
  ROUND(SUM(size_bytes) / 1048576.0, 2) AS total_mb
FROM messages
GROUP BY month
ORDER BY month;

-- Size distribution buckets
SELECT
  CASE
    WHEN size_bytes < 10000         THEN '< 10KB'
    WHEN size_bytes < 100000        THEN '10KB–100KB'
    WHEN size_bytes < 1000000       THEN '100KB–1MB'
    WHEN size_bytes >= 1000000      THEN '> 1MB'
  END AS size_bucket,
  COUNT(*) AS msg_count
FROM messages
GROUP BY size_bucket
ORDER BY MIN(size_bytes);

-- Attachment distribution
SELECT
  has_attachments,
  COUNT(*) AS msg_count
FROM messages
GROUP BY has_attachments;

-- Flag distribution
SELECT
  flags_json,
  COUNT(*) AS msg_count
FROM messages
GROUP BY flags_json
ORDER BY msg_count DESC
LIMIT 20;
