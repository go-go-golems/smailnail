-- 28-investigate-annotation-coverage.sql
-- Check annotation coverage at various stages during the triage process

-- Annotations by tag
SELECT tag, COUNT(*) as cnt FROM annotations GROUP BY tag ORDER BY cnt DESC;

-- Total annotated
SELECT COUNT(*) as total_annotations FROM annotations;

-- Messages with annotated sender
SELECT COUNT(*) as covered_msgs FROM messages m
WHERE EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=m.sender_email);

-- Total messages for percentage
SELECT COUNT(*) as total_msgs FROM messages;

-- Top unannotated senders (by message volume)
SELECT m.sender_email, m.sender_domain, COUNT(*) as cnt
FROM messages m
WHERE m.sender_email != ''
  AND NOT EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=m.sender_email)
GROUP BY m.sender_email
ORDER BY cnt DESC
LIMIT 30;
