-- 41-investigate-subject-lengths.sql
-- Subject line length distribution and distinct counts per category

-- Length buckets
SELECT 
  CASE 
    WHEN length(subject) < 30 THEN '<30 chars'
    WHEN length(subject) < 60 THEN '30-59 chars'
    WHEN length(subject) < 100 THEN '60-99 chars'
    ELSE '100+ chars'
  END as len_bucket,
  COUNT(*) as cnt
FROM messages
GROUP BY len_bucket
ORDER BY MIN(length(subject));

-- Total distinct subjects
SELECT COUNT(DISTINCT subject) as total_distinct_subjects FROM messages;

-- Distinct subjects per annotation category
SELECT 
  CASE 
    WHEN a.tag LIKE 'important/%' THEN 'IMPORTANT'
    WHEN a.tag = 'personal' THEN 'PERSONAL'
    WHEN a.tag = 'work' THEN 'WORK'
    WHEN a.tag = 'community' THEN 'COMMUNITY'
    WHEN a.tag LIKE 'newsletter/%' THEN 'NEWSLETTER'
    WHEN a.tag LIKE 'noise/%' THEN 'NOISE'
    WHEN a.tag IN ('hobby','services','financial') THEN 'OTHER_SIGNAL'
    ELSE 'UNCATEGORIZED'
  END as category,
  COUNT(DISTINCT m.subject) as unique_subjects,
  COUNT(m.id) as total_msgs
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email
GROUP BY category
ORDER BY unique_subjects DESC;
