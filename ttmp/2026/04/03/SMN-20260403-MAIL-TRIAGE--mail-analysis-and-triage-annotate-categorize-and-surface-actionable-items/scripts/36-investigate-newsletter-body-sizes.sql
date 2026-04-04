-- 36-investigate-newsletter-body-sizes.sql
-- Newsletter body text sizes and sample content (for embedding strategy)

-- Size stats by newsletter type
SELECT a.tag, COUNT(m.id) as msgs,
       ROUND(AVG(length(m.body_text))) as avg_body,
       ROUND(MIN(length(m.body_text))) as min_body,
       ROUND(MAX(length(m.body_text))) as max_body
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email
WHERE a.tag LIKE 'newsletter/%'
GROUP BY a.tag;

-- Sample newsletter body (simonw)
SELECT substr(body_text, 1, 500)
FROM messages
WHERE sender_email = 'simonw@substack.com'
ORDER BY internal_date DESC LIMIT 1;

-- Sample personal email body
SELECT substr(body_text, 1, 500)
FROM messages
WHERE sender_email = 'fahree@gmail.com'
ORDER BY internal_date DESC LIMIT 1;

-- Sample CPA email body
SELECT substr(body_text, 1, 500)
FROM messages
WHERE sender_email = 'liam@davemillercpa.com'
ORDER BY internal_date DESC LIMIT 1;
