-- 30-investigate-tax-emails.sql
-- Search for tax-related messages across all senders

-- Broad subject search
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE subject LIKE '%tax%' OR subject LIKE '%Tax%' OR subject LIKE '%1099%'
   OR subject LIKE '%W-2%' OR subject LIKE '%W2%' OR subject LIKE '%CPA%'
   OR subject LIKE '%filing%' OR subject LIKE '%IRS%' OR subject LIKE '%Steuer%'
ORDER BY internal_date DESC
LIMIT 40;

-- Specific CPA senders
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE sender_email LIKE '%davemiller%' OR sender_email LIKE '%cpa%'
   OR sender_email LIKE '%intuit%' OR sender_email LIKE '%turbotax%'
   OR sender_domain = 'davemillercpa.com'
ORDER BY internal_date DESC;
