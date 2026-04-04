-- 18-investigate-all-domains.sql
-- All unique sender domains ranked by volume

SELECT sender_domain, COUNT(*) as cnt
FROM messages
WHERE sender_domain != ''
GROUP BY sender_domain
ORDER BY cnt DESC;

-- Total unique domains
SELECT COUNT(DISTINCT sender_domain) as unique_domains FROM messages WHERE sender_domain != '';
