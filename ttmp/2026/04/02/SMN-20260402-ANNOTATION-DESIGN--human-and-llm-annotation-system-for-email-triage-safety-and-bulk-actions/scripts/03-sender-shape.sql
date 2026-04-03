.print 'Top senders by message count'
SELECT
  email,
  display_name,
  domain,
  is_private_relay,
  relay_display_domain,
  msg_count,
  first_seen_date,
  last_seen_date
FROM senders
ORDER BY msg_count DESC, email
LIMIT 50;

.print ''
.print 'Top domains'
SELECT
  domain,
  COUNT(*) AS sender_count,
  SUM(msg_count) AS total_messages
FROM senders
GROUP BY domain
ORDER BY total_messages DESC, sender_count DESC
LIMIT 50;

.print ''
.print 'Private relay senders'
SELECT
  email,
  display_name,
  relay_display_domain,
  msg_count
FROM senders
WHERE is_private_relay = 1
ORDER BY msg_count DESC, email
LIMIT 50;
