-- 40-investigate-embedding-candidates.sql
-- Queries to estimate embedding budget and identify what to embed

-- Thread counts by sender category
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
  COUNT(DISTINCT m.thread_id) as threads,
  COUNT(DISTINCT m.id) as msgs
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email
WHERE m.thread_id != ''
GROUP BY category
ORDER BY threads DESC;

-- Unique subjects in personal+important (candidate for subject-line embedding)
SELECT COUNT(DISTINCT subject) as unique_subjects FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email
WHERE a.tag = 'personal' OR a.tag LIKE 'important/%';

-- Distinct newsletter senders (candidate for sender profile embedding)
SELECT COUNT(DISTINCT target_id) as newsletter_senders FROM annotations WHERE tag LIKE 'newsletter/%';

-- Multi-message non-noise threads (candidate for thread summary embedding)
SELECT 
  CASE 
    WHEN t.message_count BETWEEN 2 AND 5 THEN '2-5'
    WHEN t.message_count BETWEEN 6 AND 20 THEN '6-20'
    WHEN t.message_count > 20 THEN '20+'
  END as size,
  COUNT(*) as threads
FROM threads t
WHERE t.message_count >= 2
  AND EXISTS (
    SELECT 1 FROM messages m WHERE m.thread_id = t.thread_id
    AND NOT EXISTS (
      SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=m.sender_email AND a.tag LIKE 'noise/%'
    )
  )
GROUP BY size;
