-- 44-investigate-thread-summarization.sql
-- Threads that would benefit from LLM summarization before embedding

-- equity/taxes thread as example: per-message sizes
SELECT substr(m.internal_date,1,10) as date, m.sender_email, 
       substr(m.subject,1,60) as subject,
       length(m.body_text) as body_len
FROM messages m
WHERE m.thread_id = (SELECT thread_id FROM threads WHERE subject = 'equity/taxes advice' LIMIT 1)
ORDER BY m.internal_date ASC;

-- Total body chars in that thread
SELECT SUM(length(body_text)) as total_body_chars
FROM messages m
WHERE m.thread_id = (SELECT thread_id FROM threads WHERE subject = 'equity/taxes advice' LIMIT 1);

-- Newsletter issue counts per sender (how many distinct issues)
SELECT sender_email, 
       COUNT(*) as msgs,
       COUNT(DISTINCT date(internal_date)) as distinct_days
FROM messages
WHERE sender_email IN ('simonw@substack.com','pragmaticengineer@substack.com','garymarcus@substack.com','hello@every.to')
GROUP BY sender_email;
