-- 17-investigate-category-breakdown.sql
-- Approximate category breakdown by domain (pre-annotation heuristic)

SELECT 'GitHub notifications' as category, COUNT(*) as cnt FROM messages WHERE sender_domain = 'github.com'
UNION ALL
SELECT 'Newsletters (substack+beehiiv+every.to+readwise)', COUNT(*) FROM messages WHERE sender_domain IN ('substack.com','mail.beehiiv.com','every.to','readwise.io')
UNION ALL
SELECT 'Commerce (amazon,thetreecenter,walgreens,cvs)', COUNT(*) FROM messages WHERE sender_domain IN ('amazon.com','thetreecenter.com','eml.walgreens.com','your.cvs.com','updates.itch.io')
UNION ALL
SELECT 'Personal (gmail,privaterelay)', COUNT(*) FROM messages WHERE sender_domain IN ('gmail.com','privaterelay.appleid.com')
UNION ALL
SELECT 'Mailing lists (entropia)', COUNT(*) FROM messages WHERE sender_domain = 'lists.entropia.de'
UNION ALL
SELECT 'Financial (paypal,bofa)', COUNT(*) FROM messages WHERE sender_domain IN ('paypal.com','ealerts.bankofamerica.com')
UNION ALL
SELECT 'Social (twitch,soundcloud,bandcamp,facebook,meetup)', COUNT(*) FROM messages WHERE sender_domain IN ('twitch.tv','notifications.soundcloud.com','bandcamp.com','facebookmail.com','email.meetup.com');
