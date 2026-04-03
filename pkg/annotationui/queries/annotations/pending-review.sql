-- Show annotations still pending review
SELECT
  id,
  target_type,
  target_id,
  tag,
  source_kind,
  source_label,
  agent_run_id,
  created_at
FROM annotations
WHERE review_state = 'to_review'
ORDER BY created_at DESC, id DESC
LIMIT 200;
