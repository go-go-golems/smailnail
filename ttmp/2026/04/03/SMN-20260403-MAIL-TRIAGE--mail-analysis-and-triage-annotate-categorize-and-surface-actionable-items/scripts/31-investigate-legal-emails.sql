-- 31-investigate-legal-emails.sql
-- Search for lawyer/legal correspondence

-- Broad subject search
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE subject LIKE '%lawyer%' OR subject LIKE '%attorney%' OR subject LIKE '%legal%'
   OR subject LIKE '%contract%' OR subject LIKE '%Anwalt%' OR subject LIKE '%Rechts%'
   OR subject LIKE '%visa%' OR subject LIKE '%immigration%' OR subject LIKE '%green card%'
   OR subject LIKE '%USCIS%' OR subject LIKE '%petition%'
ORDER BY internal_date DESC
LIMIT 30;

-- Specific law firm senders
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE sender_email LIKE '%wrightfamily%' OR sender_email LIKE '%wright%law%'
ORDER BY internal_date DESC;

-- Kryzak law
SELECT id, substr(internal_date,1,10) as date, sender_email, substr(subject,1,80) as subject
FROM messages
WHERE sender_email LIKE '%kryzak%'
ORDER BY internal_date DESC;
