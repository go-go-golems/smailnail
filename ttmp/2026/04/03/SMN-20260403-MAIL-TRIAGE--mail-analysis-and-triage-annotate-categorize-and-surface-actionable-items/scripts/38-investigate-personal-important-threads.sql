-- 38-investigate-personal-important-threads.sql
-- Multi-message threads involving personal or important senders

SELECT t.thread_id, t.subject, t.message_count, t.participant_count
FROM threads t
WHERE t.message_count >= 3
  AND EXISTS (
    SELECT 1 FROM messages m
    JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email
    WHERE m.thread_id = t.thread_id
      AND (a.tag = 'personal' OR a.tag LIKE 'important/%')
  )
ORDER BY t.message_count DESC
LIMIT 15;
