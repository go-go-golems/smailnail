-- 32-investigate-equity-emails.sql
-- Search for equity/stock option/compensation correspondence

-- Broad subject search
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE (subject LIKE '%equity%' OR subject LIKE '%offer%' OR subject LIKE '%compensation%'
   OR subject LIKE '%salary%' OR subject LIKE '%stock option%' OR subject LIKE '%vesting%'
   OR subject LIKE '%carta%' OR subject LIKE '%exercise%')
  AND sender_domain NOT IN ('github.com','twitch.tv','facebookmail.com','email.meetup.com',
      'thetreecenter.com','eml.walgreens.com','your.cvs.com','amazon.com')
  AND subject NOT LIKE '%Off%' AND subject NOT LIKE '%off%' AND subject NOT LIKE '%CI%'
ORDER BY internal_date DESC
LIMIT 30;

-- Specific equity platform senders
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE sender_domain IN ('mail.hiive.com','equitybee.com','equityzen.com')
ORDER BY internal_date DESC;

-- Admoove / bl0rg compensation
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE sender_email LIKE '%admoove%' OR sender_email LIKE '%bl0rg%' OR sender_domain = 'admoove.com'
ORDER BY internal_date DESC;
