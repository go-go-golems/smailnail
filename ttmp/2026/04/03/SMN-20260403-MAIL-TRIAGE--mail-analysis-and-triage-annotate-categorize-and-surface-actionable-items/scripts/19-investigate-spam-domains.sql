-- 19-investigate-spam-domains.sql
-- High-volume domains with unsubscribe headers (spam/marketing candidates)

SELECT s.domain, COUNT(DISTINCT s.email) as senders, SUM(s.msg_count) as total_msgs,
       MAX(s.has_list_unsubscribe) as has_unsub
FROM senders s
WHERE s.domain != ''
GROUP BY s.domain
HAVING SUM(s.msg_count) > 40
ORDER BY total_msgs DESC
LIMIT 50;
