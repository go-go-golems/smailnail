-- 16-investigate-empty-senders.sql
-- Messages with empty sender_email — enrichment gap

SELECT COUNT(*) as empty_sender_count FROM messages WHERE sender_email = '';

-- Sample: what are they? (mostly GitHub bot notifications)
SELECT id, substr(internal_date,1,16) as date, from_summary, substr(subject,1,60) as subject
FROM messages
WHERE sender_email = ''
ORDER BY internal_date DESC
LIMIT 15;
