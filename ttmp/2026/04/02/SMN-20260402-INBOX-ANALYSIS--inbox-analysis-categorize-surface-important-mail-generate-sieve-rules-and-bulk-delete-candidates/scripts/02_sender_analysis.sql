-- 02_sender_analysis.sql
-- Sender breakdown: top senders, domain analysis

-- Top 60 senders by volume
SELECT
  from_summary,
  COUNT(*) AS msg_count,
  ROUND(SUM(size_bytes)/1024.0, 0) AS total_kb
FROM messages
GROUP BY from_summary
ORDER BY msg_count DESC
LIMIT 60;

-- Top sender domains extracted from from_summary
SELECT
  LOWER(
    CASE
      WHEN instr(from_summary, '@') > 0
      THEN substr(from_summary,
             instr(from_summary, '@') + 1,
             CASE
               WHEN instr(substr(from_summary, instr(from_summary,'@')+1), '>') > 0
               THEN instr(substr(from_summary, instr(from_summary,'@')+1), '>') - 1
               WHEN instr(substr(from_summary, instr(from_summary,'@')+1), ' ') > 0
               THEN instr(substr(from_summary, instr(from_summary,'@')+1), ' ') - 1
               ELSE length(from_summary)
             END
           )
      ELSE from_summary
    END
  ) AS domain,
  COUNT(*) AS msg_count
FROM messages
GROUP BY domain
ORDER BY msg_count DESC
LIMIT 60;
