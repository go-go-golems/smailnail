-- 45b-investigate-monthly-digest-fast.sql
-- Monthly signal counts via pre-aggregated sender counts (avoids full messages scan)
-- The join to messages is slow due to no index on sender_email.
-- Workaround: count messages per sender_email first, then join to annotations.

WITH sender_months AS (
  SELECT sender_email, strftime('%Y-%m', internal_date) as month, COUNT(*) as cnt
  FROM messages
  WHERE sender_email != ''
  GROUP BY sender_email, month
)
SELECT sm.month,
  SUM(CASE WHEN a.tag = 'personal' OR a.tag LIKE 'important/%' THEN sm.cnt ELSE 0 END) as important,
  SUM(CASE WHEN a.tag LIKE 'newsletter/%' THEN sm.cnt ELSE 0 END) as newsletters,
  SUM(CASE WHEN a.tag = 'work' THEN sm.cnt ELSE 0 END) as work,
  SUM(CASE WHEN a.tag = 'community' THEN sm.cnt ELSE 0 END) as community,
  SUM(CASE WHEN a.tag IN ('hobby','services','financial') THEN sm.cnt ELSE 0 END) as other_signal
FROM sender_months sm
JOIN annotations a ON a.target_type='sender' AND a.target_id=sm.sender_email
WHERE a.tag NOT LIKE 'noise/%'
GROUP BY sm.month
ORDER BY sm.month DESC
LIMIT 12;
