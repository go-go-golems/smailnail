-- 45-investigate-monthly-digest.sql
-- Monthly message counts per signal category (for time-windowed embedding)

SELECT strftime('%Y-%m', m.internal_date) as month,
  SUM(CASE WHEN a.tag = 'personal' OR a.tag LIKE 'important/%' THEN 1 ELSE 0 END) as important,
  SUM(CASE WHEN a.tag LIKE 'newsletter/%' THEN 1 ELSE 0 END) as newsletters,
  SUM(CASE WHEN a.tag = 'work' THEN 1 ELSE 0 END) as work,
  SUM(CASE WHEN a.tag = 'community' THEN 1 ELSE 0 END) as community,
  SUM(CASE WHEN a.tag IN ('hobby','services','financial') THEN 1 ELSE 0 END) as other_signal
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email
WHERE a.tag NOT LIKE 'noise/%'
GROUP BY month
ORDER BY month DESC
LIMIT 12;
