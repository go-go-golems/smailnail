-- 33-investigate-conferences-talks.sql
-- Search for conference/speaking/talk invitations

-- Broad subject search (excluding noisy senders)
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE (subject LIKE '%speak%' OR subject LIKE '%talk%' OR subject LIKE '%conference%'
   OR subject LIKE '%keynote%' OR subject LIKE '%invitation%' OR subject LIKE '%invite%'
   OR subject LIKE '%Vortrag%' OR subject LIKE '%Einladung%' OR subject LIKE '%CFP%'
   OR subject LIKE '%call for%' OR subject LIKE '%presentation%' OR subject LIKE '%panel%'
   OR subject LIKE '%workshop%')
  AND sender_domain NOT IN ('github.com','twitch.tv','facebookmail.com','email.meetup.com')
  AND subject NOT LIKE '%CI%' AND subject NOT LIKE '%PR run%' AND subject NOT LIKE '%Run failed%'
ORDER BY internal_date DESC
LIMIT 40;

-- Conference-specific senders
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE sender_domain IN ('ai.engineer','mlops.community','courses.maven.com')
   OR sender_email LIKE '%conference%'
   OR sender_email LIKE '%demetrios%'
ORDER BY internal_date DESC
LIMIT 20;
