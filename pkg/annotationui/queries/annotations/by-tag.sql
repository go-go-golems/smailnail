-- Count annotations grouped by tag
SELECT
  tag,
  COUNT(*) AS count
FROM annotations
GROUP BY tag
ORDER BY count DESC, tag ASC;
