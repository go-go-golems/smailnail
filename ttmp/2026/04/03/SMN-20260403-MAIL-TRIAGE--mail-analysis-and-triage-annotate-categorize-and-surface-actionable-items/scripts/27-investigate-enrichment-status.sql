-- 27-investigate-enrichment-status.sql
-- Check completeness of enrichment fields

SELECT 'thread_id populated' as metric, COUNT(*) as value FROM messages WHERE thread_id != ''
UNION ALL
SELECT 'sender_email populated', COUNT(*) FROM messages WHERE sender_email != ''
UNION ALL
SELECT 'body_text populated', COUNT(*) FROM messages WHERE body_text != ''
UNION ALL
SELECT 'body_html populated', COUNT(*) FROM messages WHERE body_html != '';

-- Sample body_text
SELECT substr(body_text, 1, 300) FROM messages WHERE body_text != '' ORDER BY internal_date DESC LIMIT 1;
