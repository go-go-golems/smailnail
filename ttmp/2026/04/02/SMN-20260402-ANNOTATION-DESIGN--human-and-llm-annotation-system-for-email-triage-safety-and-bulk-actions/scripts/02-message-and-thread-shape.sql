.print 'Message coverage'
SELECT
  COUNT(*) AS total_messages,
  SUM(CASE WHEN remote_deleted THEN 1 ELSE 0 END) AS remote_deleted_messages,
  SUM(CASE WHEN has_attachments THEN 1 ELSE 0 END) AS attachment_messages,
  SUM(CASE WHEN thread_id != '' THEN 1 ELSE 0 END) AS threaded_messages,
  SUM(CASE WHEN json_extract(headers_json, '$.References') IS NOT NULL
             OR json_extract(headers_json, '$."In-Reply-To"') IS NOT NULL
      THEN 1 ELSE 0 END) AS messages_with_reply_headers
FROM messages;

.print ''
.print 'Thread size distribution'
SELECT message_count, COUNT(*) AS threads
FROM threads
GROUP BY message_count
ORDER BY message_count;

.print ''
.print 'Largest threads'
SELECT
  thread_id,
  message_count,
  participant_count,
  subject,
  account_key,
  mailbox_name,
  first_sent_date,
  last_sent_date
FROM threads
ORDER BY message_count DESC, last_sent_date DESC
LIMIT 25;

.print ''
.print 'Singleton thread sample'
SELECT
  thread_id,
  subject,
  from_summary,
  sent_date
FROM messages
WHERE thread_id != ''
GROUP BY thread_id
HAVING COUNT(*) = 1
ORDER BY sent_date DESC
LIMIT 25;
