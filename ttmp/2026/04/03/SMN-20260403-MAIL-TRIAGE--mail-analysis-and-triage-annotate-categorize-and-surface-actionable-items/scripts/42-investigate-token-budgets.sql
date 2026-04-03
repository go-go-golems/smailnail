-- 42-investigate-token-budgets.sql
-- Estimate token budgets for different embedding strategies
-- Rough token estimate: chars / 4

-- Subject lines (all non-noise)
SELECT 'non-noise subjects' as what,
  SUM(length(subject)) as total_chars,
  COUNT(*) as msgs,
  ROUND(SUM(length(subject)) / 4.0) as est_tokens
FROM messages m
WHERE EXISTS (
  SELECT 1 FROM annotations a 
  WHERE a.target_type='sender' AND a.target_id=m.sender_email AND a.tag NOT LIKE 'noise/%'
);

-- Body text of personal+important only
SELECT 'personal+important bodies' as what,
  SUM(length(body_text)) as total_chars,
  COUNT(*) as msgs,
  ROUND(SUM(length(body_text)) / 4.0) as est_tokens
FROM messages m
WHERE EXISTS (
  SELECT 1 FROM annotations a 
  WHERE a.target_type='sender' AND a.target_id=m.sender_email 
  AND (a.tag = 'personal' OR a.tag LIKE 'important/%')
);

-- Body text of newsletters
SELECT 'newsletter bodies' as what,
  SUM(length(body_text)) as total_chars,
  COUNT(*) as msgs,
  ROUND(SUM(length(body_text)) / 4.0) as est_tokens
FROM messages m
WHERE EXISTS (
  SELECT 1 FROM annotations a 
  WHERE a.target_type='sender' AND a.target_id=m.sender_email AND a.tag LIKE 'newsletter/%'
);
