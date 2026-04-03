-- 48-investigate-query-reuse-patterns.sql
-- Identifying common query patterns that could be embedded as reusable tools
-- These are the "query archetypes" we used repeatedly in this session

-- ARCHETYPE 1: "Show me recent messages from [category]"
-- Used for: personal, work, community, financial, important/*
-- Template:
--   SELECT id, date, sender_email, subject FROM messages m
--   JOIN annotations a ON ... WHERE a.tag = ? AND m.internal_date >= date('now', '-N days')

-- ARCHETYPE 2: "Who are the top senders in [category]?"
-- Used for: coverage checking, unsubscribe candidates
-- Template:
--   SELECT sender_email, COUNT(*) FROM messages m
--   JOIN annotations a ON ... WHERE a.tag LIKE ? GROUP BY sender_email ORDER BY COUNT(*) DESC

-- ARCHETYPE 3: "Find messages about [topic]"
-- Used for: tax, legal, equity, conferences
-- Template:
--   SELECT id, date, sender_email, subject FROM messages
--   WHERE subject LIKE '%keyword%' ORDER BY internal_date DESC

-- ARCHETYPE 4: "What threads have real conversations?"
-- Used for: personal threads, important threads
-- Template:
--   SELECT thread_id, subject, message_count FROM threads
--   WHERE message_count >= N AND EXISTS (sender in category)

-- ARCHETYPE 5: "Monthly breakdown by category"
-- Used for: digest generation, trend analysis
-- Template:
--   WITH sender_months AS (...) SELECT month, SUM(CASE...) FROM ... GROUP BY month

-- How many scripts did we create? (meta-query)
-- 00-08: action scripts, 09-48: investigation queries
-- Total: 49 files covering the full investigation trail
