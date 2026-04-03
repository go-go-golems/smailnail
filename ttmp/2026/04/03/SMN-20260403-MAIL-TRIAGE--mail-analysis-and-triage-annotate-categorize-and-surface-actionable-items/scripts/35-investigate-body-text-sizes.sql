-- 35-investigate-body-text-sizes.sql
-- Body text size distribution (for embedding budget estimation)

SELECT 
  CASE 
    WHEN length(body_text) = 0 THEN '0 (empty)'
    WHEN length(body_text) < 200 THEN '1-199 chars'
    WHEN length(body_text) < 1000 THEN '200-999 chars'
    WHEN length(body_text) < 5000 THEN '1K-5K chars'
    WHEN length(body_text) < 20000 THEN '5K-20K chars'
    ELSE '20K+ chars'
  END as size_bucket,
  COUNT(*) as cnt,
  ROUND(AVG(length(body_text))) as avg_len
FROM messages
GROUP BY size_bucket
ORDER BY MIN(length(body_text));

-- Important sender message counts
SELECT a.tag, COUNT(DISTINCT m.id) as msgs
FROM annotations a
JOIN messages m ON m.sender_email = a.target_id
WHERE a.target_type = 'sender' AND a.tag LIKE 'important/%'
GROUP BY a.tag ORDER BY msgs DESC;

-- Personal sender body text sizes
SELECT m.sender_email, COUNT(*) as msgs, ROUND(AVG(length(m.body_text))) as avg_body_len
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email AND a.tag='personal'
GROUP BY m.sender_email
ORDER BY msgs DESC LIMIT 15;
