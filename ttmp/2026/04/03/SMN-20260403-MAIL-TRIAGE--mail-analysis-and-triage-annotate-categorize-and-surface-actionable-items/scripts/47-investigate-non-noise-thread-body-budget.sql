-- 47-investigate-non-noise-thread-body-budget.sql
-- For multi-message non-noise threads: total body text to summarize

-- Use CTE to pre-filter senders, then join
WITH noise_senders AS (
  SELECT target_id as email FROM annotations 
  WHERE target_type='sender' AND tag LIKE 'noise/%'
),
signal_senders AS (
  SELECT target_id as email FROM annotations 
  WHERE target_type='sender' AND tag NOT LIKE 'noise/%'
)
SELECT 
  CASE 
    WHEN t.message_count BETWEEN 2 AND 5 THEN '2-5 msgs'
    WHEN t.message_count BETWEEN 6 AND 20 THEN '6-20 msgs'
    WHEN t.message_count > 20 THEN '20+ msgs'
  END as thread_size,
  COUNT(DISTINCT t.thread_id) as threads
FROM threads t
WHERE t.message_count >= 2
  AND EXISTS (
    SELECT 1 FROM messages m 
    WHERE m.thread_id = t.thread_id
    AND m.sender_email IN (SELECT email FROM signal_senders)
  )
GROUP BY thread_size;

-- Count of unique senders per multi-msg signal thread (to gauge conversation-ness)
SELECT 
  t.subject, t.message_count, t.participant_count,
  substr(t.last_sent_date,1,10) as last_date
FROM threads t
WHERE t.message_count >= 3
  AND t.participant_count >= 2
  AND EXISTS (
    SELECT 1 FROM messages m 
    JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email
    WHERE m.thread_id = t.thread_id
    AND (a.tag = 'personal' OR a.tag LIKE 'important/%' OR a.tag = 'community')
  )
ORDER BY t.last_sent_date DESC
LIMIT 20;
