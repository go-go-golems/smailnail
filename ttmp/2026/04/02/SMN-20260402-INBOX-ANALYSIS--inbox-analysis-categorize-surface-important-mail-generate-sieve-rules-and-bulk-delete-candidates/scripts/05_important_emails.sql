-- 05_important_emails.sql
-- Surface likely-important / human emails

-- 1. Direct personal (known humans / personal domains)
SELECT 'personal' AS bucket, sent_date, from_summary, subject
FROM messages
WHERE from_summary LIKE '%gmail.com%'
   OR from_summary LIKE '%@bl0rg.net%'
   OR from_summary LIKE '%goldhamstermoegen@icloud.com%'  -- Manfred (family?)
   OR from_summary LIKE '%icloud.com%'
   OR from_summary LIKE '%hotmail.com%'
   OR from_summary LIKE '%wanadoo.fr%'
   OR from_summary LIKE '%unimib.it%'
ORDER BY sent_date DESC;

-- 2. Transactional/billing (invoices, receipts, legal)
SELECT 'billing' AS bucket, sent_date, from_summary, subject
FROM messages
WHERE from_summary LIKE '%hetzner.com%'
   OR from_summary LIKE '%stripe.com%'
   OR from_summary LIKE '%invoice%'
   OR from_summary LIKE '%invoicing.co%'
   OR from_summary LIKE '%lendingclub.com%'
   OR from_summary LIKE '%venmo.com%'
   OR from_summary LIKE '%wrightfamilylawgroup.com%'
   OR subject LIKE '%invoice%'
   OR subject LIKE '%receipt%'
   OR subject LIKE '%bill%'
   OR subject LIKE '%payment%'
   OR subject LIKE '%statement%'
ORDER BY sent_date DESC;

-- 3. Developer / work tools (Rayobyte, CodeRabbit, OpenHands, W&B, Rollbar)
SELECT 'dev-tools' AS bucket, sent_date, from_summary, subject
FROM messages
WHERE from_summary LIKE '%rayobyte.com%'
   OR from_summary LIKE '%coderabbit.ai%'
   OR from_summary LIKE '%all-hands.dev%'
   OR from_summary LIKE '%wandb.ai%'
   OR from_summary LIKE '%rollbar.com%'
   OR from_summary LIKE '%augmentcode.com%'
   OR from_summary LIKE '%replit.com%'
   OR from_summary LIKE '%firecrawl.dev%'
   OR from_summary LIKE '%omi.me%'
   OR from_summary LIKE '%mermaid.ai%'
   OR from_summary LIKE '%brightdata.com%'
ORDER BY sent_date DESC;

-- 4. Apple / Microsoft account (security but possibly important)
SELECT 'account-mgmt' AS bucket, sent_date, from_summary, subject
FROM messages
WHERE from_summary LIKE '%apple.com%'
   OR from_summary LIKE '%applecard.apple%'
   OR from_summary LIKE '%insideapple.apple.com%'
   OR from_summary LIKE '%microsoft.com%'
   OR from_summary LIKE '%accountprotection.microsoft.com%'
   OR from_summary LIKE '%accounts.google.com%'
ORDER BY sent_date DESC;
