-- 12-investigate-messages-per-month.sql
-- Message volume by month to understand data completeness

SELECT strftime('%Y-%m', internal_date) as month, COUNT(*) as cnt
FROM messages
GROUP BY month
ORDER BY month;
