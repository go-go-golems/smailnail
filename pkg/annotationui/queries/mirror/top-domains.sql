-- Rank sender domains by message volume
SELECT
  domain,
  COUNT(*) AS sender_count,
  SUM(msg_count) AS message_count
FROM senders
GROUP BY domain
ORDER BY message_count DESC, domain ASC
LIMIT 100;
