-- 07_github_breakdown.sql
-- Break down 771 GitHub messages by repo/type

-- Top repos by volume
SELECT
  CASE
    WHEN subject LIKE '%[wesen/%' THEN regexp_substr(subject, '\[wesen/[^\]]+\]')
    WHEN subject LIKE '%[%/%]%'   THEN substr(subject, instr(subject,'[')+1, instr(subject,']')-instr(subject,'[')-1)
    ELSE '(other)'
  END AS repo,
  COUNT(*) AS cnt
FROM messages
WHERE from_summary LIKE '%notifications@github.com%'
   OR from_summary LIKE '%noreply@github.com%'
GROUP BY repo
ORDER BY cnt DESC
LIMIT 30;

-- GitHub notification types from subject lines
SELECT
  CASE
    WHEN subject LIKE 'Re: [%'          THEN 'discussion-reply'
    WHEN subject LIKE '[%] [%'          THEN 'label-event'
    WHEN subject LIKE '%merged%'         THEN 'PR-merged'
    WHEN subject LIKE '%closed%'         THEN 'closed'
    WHEN subject LIKE '%opened%'         THEN 'opened'
    WHEN subject LIKE '%commented%'      THEN 'comment'
    WHEN subject LIKE '%review%'         THEN 'review'
    WHEN subject LIKE '%approved%'       THEN 'approved'
    WHEN subject LIKE '%assigned%'       THEN 'assigned'
    WHEN subject LIKE '%pushed%'         THEN 'push'
    WHEN subject LIKE '%[GitHub]%'       THEN 'github-system'
    ELSE 'other'
  END AS gh_type,
  COUNT(*) AS cnt
FROM messages
WHERE from_summary LIKE '%notifications@github.com%'
   OR from_summary LIKE '%noreply@github.com%'
GROUP BY gh_type
ORDER BY cnt DESC;

-- Codex bot volume
SELECT COUNT(*) AS codex_bot_msgs FROM messages
WHERE from_summary LIKE '%chatgpt-codex-connector%';

-- Personal GH notifications (not Codex bot)
SELECT COUNT(*) AS personal_gh FROM messages
WHERE from_summary LIKE '%Manuel Odendahl%notifications@github.com%';
