-- 14-investigate-substack-newsletters.sql
-- Breakdown of all Substack newsletter senders

SELECT sender_email, COUNT(*) as cnt
FROM messages
WHERE sender_domain = 'substack.com'
GROUP BY sender_email
ORDER BY cnt DESC
LIMIT 30;
