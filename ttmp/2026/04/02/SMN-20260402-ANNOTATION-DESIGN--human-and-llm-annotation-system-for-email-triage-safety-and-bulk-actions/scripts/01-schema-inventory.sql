.print 'Mirror metadata'
SELECT key, value
FROM mirror_metadata
WHERE key IN ('schema_version', 'fts5_status', 'enrich_senders_at', 'enrich_threads_at', 'enrich_unsubscribe_at')
ORDER BY key;

.print ''
.print 'Core table row counts'
SELECT 'messages' AS table_name, COUNT(*) AS row_count FROM messages
UNION ALL
SELECT 'senders', COUNT(*) FROM senders
UNION ALL
SELECT 'threads', COUNT(*) FROM threads
ORDER BY table_name;

.print ''
.print 'Messages columns'
PRAGMA table_info(messages);

.print ''
.print 'Senders columns'
PRAGMA table_info(senders);

.print ''
.print 'Threads columns'
PRAGMA table_info(threads);
