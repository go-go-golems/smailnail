-- 22-investigate-personal-senders.sql
-- Personal email senders from consumer domains

SELECT sender_email, COUNT(*) as cnt
FROM messages
WHERE sender_domain IN ('gmail.com','yahoo.com','hotmail.com','outlook.com','proton.me','protonmail.com','icloud.com')
  AND sender_email != ''
GROUP BY sender_email
ORDER BY cnt DESC
LIMIT 25;
