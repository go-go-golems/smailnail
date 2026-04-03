-- 29-investigate-group-membership.sql
-- Verify group membership after creation

-- Group list with member counts
SELECT g.name, COUNT(m.target_id) as members
FROM target_groups g
JOIN target_group_members m ON m.group_id = g.id
GROUP BY g.id
ORDER BY members DESC;

-- Direct member count
SELECT COUNT(*) as total_group_members FROM target_group_members;

-- Sample members
SELECT * FROM target_group_members LIMIT 5;

-- Group details
SELECT * FROM target_groups;

-- Inspect one group and its members
-- SELECT g.id, g.name, m.target_type, m.target_id, m.added_at
-- FROM target_groups g
-- JOIN target_group_members m ON m.group_id = g.id
-- WHERE g.name = 'Unsubscribe Candidates'
-- ORDER BY m.added_at DESC;
