-- 26-investigate-mid-tier-domains.sql
-- Domains not in top-volume list but potentially interesting

SELECT sender_domain, COUNT(*) as cnt
FROM messages
WHERE sender_domain != ''
  AND sender_domain NOT IN (
    'github.com','substack.com','privaterelay.appleid.com','thetreecenter.com',
    'email.meetup.com','amazon.com','twitch.tv','every.to','lists.entropia.de',
    'gmail.com','mail.rollbar.com','readwise.io','notifications.freelancer.com',
    'iphonephotographyschool.com','karlsruhe.stadtmobil.de','manning.com',
    'facebookmail.com','your.cvs.com','bozzuto.com'
  )
GROUP BY sender_domain
ORDER BY cnt DESC
LIMIT 30;
