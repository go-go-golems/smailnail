-- 37-investigate-category-message-counts.sql
-- Message counts by annotation category (the slow query)

-- WARNING: This query takes ~5 minutes due to SCAN on messages table.
-- See 39-investigate-slow-query-analysis.sql for analysis and faster alternatives.

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
  COUNT(DISTINCT m.id) as msgs
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email
GROUP BY category
ORDER BY msgs DESC;
