-- 07b_github_breakdown_fixed.sql
-- No regexp_substr - use LIKE + INSTR tricks only

-- Repo distribution from subjects like "[owner/repo] ..."
SELECT
  CASE
    WHEN subject LIKE '[go-go-golems/geppetto]%'           THEN 'go-go-golems/geppetto'
    WHEN subject LIKE '%[go-go-golems/geppetto]%'          THEN 'go-go-golems/geppetto'
    WHEN subject LIKE '[go-go-golems/pinocchio]%'          THEN 'go-go-golems/pinocchio'
    WHEN subject LIKE '%[go-go-golems/pinocchio]%'         THEN 'go-go-golems/pinocchio'
    WHEN subject LIKE '[go-go-golems/bobatea]%'            THEN 'go-go-golems/bobatea'
    WHEN subject LIKE '%[go-go-golems/bobatea]%'           THEN 'go-go-golems/bobatea'
    WHEN subject LIKE '[go-go-golems/codex-sessions]%'     THEN 'go-go-golems/codex-sessions'
    WHEN subject LIKE '%[go-go-golems/codex-sessions]%'    THEN 'go-go-golems/codex-sessions'
    WHEN subject LIKE '[go-go-golems/go-go-os-frontend]%'  THEN 'go-go-golems/go-go-os-frontend'
    WHEN subject LIKE '%[go-go-golems/go-go-os-frontend]%' THEN 'go-go-golems/go-go-os-frontend'
    WHEN subject LIKE '[go-go-golems/%'                    THEN 'go-go-golems/other'
    WHEN subject LIKE '%[go-go-golems/%'                   THEN 'go-go-golems/other'
    WHEN subject LIKE '[wesen/temporal-relationships]%'    THEN 'wesen/temporal-relationships'
    WHEN subject LIKE '%[wesen/temporal-relationships]%'   THEN 'wesen/temporal-relationships'
    WHEN subject LIKE '[wesen/goldeneaglecoin%'            THEN 'wesen/goldeneaglecoin.com'
    WHEN subject LIKE '%[wesen/%'                          THEN 'wesen/other'
    WHEN subject LIKE '%[GitHub]%'                         THEN 'github-system'
    ELSE 'other-gh'
  END AS repo,
  COUNT(*) AS cnt
FROM messages
WHERE from_summary LIKE '%notifications@github.com%'
   OR from_summary LIKE '%noreply@github.com%'
GROUP BY repo
ORDER BY cnt DESC;

-- GitHub CI failures vs PR discussions
SELECT
  CASE
    WHEN subject LIKE '%Run failed%'      THEN 'CI-failure'
    WHEN subject LIKE '%PR run failed%'   THEN 'CI-failure'
    WHEN subject LIKE 'Re: [%'            THEN 'PR/issue discussion'
    WHEN subject LIKE '%[GitHub] Payment%' THEN 'billing'
    WHEN subject LIKE '%[GitHub]%'        THEN 'github-system'
    ELSE 'other'
  END AS gh_event,
  COUNT(*) AS cnt
FROM messages
WHERE from_summary LIKE '%notifications@github.com%'
   OR from_summary LIKE '%noreply@github.com%'
GROUP BY gh_event
ORDER BY cnt DESC;
