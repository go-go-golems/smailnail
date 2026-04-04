-- 46-investigate-annotation-vocabulary.sql
-- The tag vocabulary and scripts inventory (for embedding scripts/queries themselves)

-- All distinct tags
SELECT DISTINCT tag FROM annotations ORDER BY tag;

-- Scripts we have (for reference — actual listing done in shell)
-- 00-08: action scripts (.sh)
-- 09-46: investigation queries (.sql)

-- Group names
SELECT name, description FROM target_groups;

-- Log entries
SELECT id, title, source_label, created_at FROM annotation_logs ORDER BY created_at;
