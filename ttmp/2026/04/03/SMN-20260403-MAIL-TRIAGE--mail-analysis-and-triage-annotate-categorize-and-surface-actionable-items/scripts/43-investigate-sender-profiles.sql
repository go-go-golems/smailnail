-- 43-investigate-sender-profiles.sql
-- Data available for synthetic sender profile generation

-- Sender metadata richness
SELECT 
  COUNT(*) as total_senders,
  SUM(CASE WHEN display_name != '' THEN 1 ELSE 0 END) as has_display_name,
  SUM(CASE WHEN has_list_unsubscribe THEN 1 ELSE 0 END) as has_unsubscribe,
  SUM(CASE WHEN msg_count >= 10 THEN 1 ELSE 0 END) as has_10plus_msgs,
  SUM(CASE WHEN msg_count >= 50 THEN 1 ELSE 0 END) as has_50plus_msgs,
  ROUND(AVG(msg_count), 1) as avg_msg_count
FROM senders;

-- Sample sender profile data
SELECT s.email, s.display_name, s.domain, s.msg_count, 
       s.first_seen_date, s.last_seen_date,
       s.has_list_unsubscribe,
       a.tag, a.note_markdown
FROM senders s
LEFT JOIN annotations a ON a.target_type='sender' AND a.target_id=s.email
WHERE s.email = 'dave@davemillercpa.com';

-- How many senders have annotations? (would produce profiles)
SELECT COUNT(*) as annotated_senders FROM senders s
WHERE EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=s.email);
