-- 23-investigate-thread-sizes.sql
-- Thread size distribution and largest threads

-- Distribution
SELECT 
  CASE 
    WHEN message_count = 1 THEN '1 msg'
    WHEN message_count BETWEEN 2 AND 5 THEN '2-5 msgs'
    WHEN message_count BETWEEN 6 AND 20 THEN '6-20 msgs'
    WHEN message_count > 20 THEN '20+ msgs'
  END as thread_size,
  COUNT(*) as num_threads
FROM threads
GROUP BY thread_size
ORDER BY num_threads DESC;

-- Largest threads (20+ messages)
SELECT thread_id, subject, message_count, participant_count,
       substr(first_sent_date,1,10) as first, substr(last_sent_date,1,10) as last
FROM threads
WHERE message_count > 20
ORDER BY message_count DESC;
