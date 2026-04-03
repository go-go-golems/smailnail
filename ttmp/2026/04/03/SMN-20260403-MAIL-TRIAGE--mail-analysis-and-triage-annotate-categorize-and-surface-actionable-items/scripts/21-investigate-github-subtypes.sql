-- 21-investigate-github-subtypes.sql
-- Break down GitHub notifications by type (CI failures, thread replies, etc.)

SELECT 
  CASE 
    WHEN subject LIKE '%PR run failed%' OR subject LIKE '%Run failed%' THEN 'CI failures'
    WHEN subject LIKE '%Dependabot%' OR subject LIKE '%dependency%' THEN 'Dependabot'
    WHEN subject LIKE '%Re: %' THEN 'Thread replies'
    WHEN subject LIKE '%Pull Request%' OR subject LIKE '%merged%' THEN 'PR activity'
    WHEN subject LIKE '%issue%' OR subject LIKE '%Issue%' THEN 'Issues'
    ELSE 'Other'
  END as category,
  COUNT(*) as cnt
FROM messages
WHERE sender_domain = 'github.com'
GROUP BY category
ORDER BY cnt DESC;
