-- 13-investigate-unsubscribe-senders.sql
-- Senders with List-Unsubscribe headers (candidates for unsubscribing)

SELECT COUNT(*) as total_with_unsubscribe FROM senders WHERE has_list_unsubscribe = 1;

-- Top domains with unsubscribe
SELECT domain, COUNT(*) as cnt
FROM senders
WHERE has_list_unsubscribe = 1
GROUP BY domain
ORDER BY cnt DESC
LIMIT 20;
