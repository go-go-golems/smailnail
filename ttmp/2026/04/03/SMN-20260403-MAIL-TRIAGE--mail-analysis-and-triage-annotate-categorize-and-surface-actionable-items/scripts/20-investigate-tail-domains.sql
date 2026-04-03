-- 20-investigate-tail-domains.sql
-- Low-volume domains (1-5 messages) — likely personal or one-off

SELECT sender_domain, COUNT(*) as cnt
FROM messages
WHERE sender_domain != ''
GROUP BY sender_domain
HAVING cnt <= 5
ORDER BY cnt DESC, sender_domain
LIMIT 40;
