.print 'Potential annotation entity counts'
SELECT 'accounts' AS entity_type, COUNT(DISTINCT account_key) AS count FROM messages
UNION ALL
SELECT 'mailboxes', COUNT(DISTINCT account_key || '::' || mailbox_name) FROM messages
UNION ALL
SELECT 'messages', COUNT(*) FROM messages
UNION ALL
SELECT 'threads', COUNT(*) FROM threads
UNION ALL
SELECT 'senders', COUNT(*) FROM senders
UNION ALL
SELECT 'domains', COUNT(DISTINCT domain) FROM senders
ORDER BY entity_type;

.print ''
.print 'Message candidates with likely manual review value'
SELECT
  m.sent_date,
  m.subject,
  m.from_summary,
  s.msg_count AS sender_message_count,
  s.has_list_unsubscribe,
  m.thread_id,
  t.message_count AS thread_message_count
FROM messages m
LEFT JOIN senders s ON s.email = m.sender_email
LEFT JOIN threads t ON t.thread_id = m.thread_id
WHERE
  s.msg_count >= 3
  OR t.message_count >= 3
  OR s.has_list_unsubscribe = 1
ORDER BY m.sent_date DESC
LIMIT 100;

.print ''
.print 'Messages with no sender normalization or threading'
SELECT
  COUNT(*) AS messages_without_sender_email
FROM messages
WHERE sender_email = '';

SELECT
  COUNT(*) AS messages_without_thread_id
FROM messages
WHERE thread_id = '';
