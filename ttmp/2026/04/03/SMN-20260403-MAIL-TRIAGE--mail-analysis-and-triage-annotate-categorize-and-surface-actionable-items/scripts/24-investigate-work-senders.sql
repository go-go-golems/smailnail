-- 24-investigate-work-senders.sql
-- Work-related senders (team-mento, slack, zulip, rollbar)

SELECT sender_email, COUNT(*) as cnt
FROM messages
WHERE sender_domain IN ('slack.com','zulip.com','mail.rollbar.com')
   OR subject LIKE '%team-mento%'
   OR subject LIKE '%mento%'
GROUP BY sender_email
ORDER BY cnt DESC
LIMIT 15;

-- Zulip subjects (Recurse Center)
SELECT substr(internal_date,1,10) as date, substr(subject,1,80) as subject
FROM messages
WHERE sender_domain = 'zulip.com'
ORDER BY internal_date DESC
LIMIT 10;
