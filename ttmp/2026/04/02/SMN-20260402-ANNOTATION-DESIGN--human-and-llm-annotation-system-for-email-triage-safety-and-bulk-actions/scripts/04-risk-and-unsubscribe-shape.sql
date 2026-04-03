.print 'Senders with unsubscribe metadata'
SELECT
  COUNT(*) AS senders_with_unsubscribe,
  SUM(CASE WHEN unsubscribe_mailto != '' THEN 1 ELSE 0 END) AS senders_with_mailto,
  SUM(CASE WHEN unsubscribe_http != '' THEN 1 ELSE 0 END) AS senders_with_http
FROM senders
WHERE has_list_unsubscribe = 1;

.print ''
.print 'Top bulk-action candidate senders'
SELECT
  email,
  display_name,
  domain,
  msg_count,
  has_list_unsubscribe,
  unsubscribe_http,
  unsubscribe_mailto
FROM senders
WHERE msg_count >= 5
ORDER BY msg_count DESC, email
LIMIT 50;

.print ''
.print 'Payment-ish and reminder-ish message sample'
SELECT
  sent_date,
  subject,
  from_summary
FROM messages
WHERE lower(subject) LIKE '%payment%'
   OR lower(subject) LIKE '%reminder%'
   OR lower(subject) LIKE '%invoice%'
ORDER BY sent_date DESC
LIMIT 50;
