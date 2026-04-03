-- 15-investigate-recent-messages.sql
-- Most recent 20 messages to understand current inbox state

SELECT id, substr(internal_date,1,16) as date, sender_email, substr(subject,1,60) as subject
FROM messages
ORDER BY internal_date DESC
LIMIT 20;
