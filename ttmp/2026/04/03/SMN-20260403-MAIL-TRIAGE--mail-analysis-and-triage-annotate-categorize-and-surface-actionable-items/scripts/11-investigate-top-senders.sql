-- 11-investigate-top-senders.sql
-- Top 40 senders by message volume

SELECT sender_email, sender_domain, COUNT(*) as cnt
FROM messages
GROUP BY sender_email
ORDER BY cnt DESC
LIMIT 40;
