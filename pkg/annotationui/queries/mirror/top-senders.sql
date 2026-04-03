-- Show senders with the highest message counts
SELECT
  email,
  display_name,
  domain,
  msg_count
FROM senders
ORDER BY msg_count DESC, email ASC
LIMIT 100;
