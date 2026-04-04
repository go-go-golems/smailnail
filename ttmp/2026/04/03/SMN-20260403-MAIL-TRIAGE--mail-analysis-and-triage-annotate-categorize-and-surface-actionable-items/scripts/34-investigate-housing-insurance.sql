-- 34-investigate-housing-insurance.sql
-- Search for housing, insurance, rent, lease correspondence

SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE (subject LIKE '%insurance%' OR subject LIKE '%Versicherung%' OR subject LIKE '%lease%'
   OR subject LIKE '%rent%' OR subject LIKE '%renewal%' OR subject LIKE '%expir%')
  AND sender_domain NOT IN ('github.com','twitch.tv','facebookmail.com','email.meetup.com',
      'thetreecenter.com','amazon.com','eml.walgreens.com','substack.com')
  AND subject NOT LIKE '%CI%' AND subject NOT LIKE '%PR run%'
ORDER BY internal_date DESC
LIMIT 25;

-- Bilt Rewards rent
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE sender_domain LIKE '%biltrewards%'
ORDER BY internal_date DESC
LIMIT 10;
