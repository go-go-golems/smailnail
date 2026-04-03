-- 25-investigate-hobby-creative.sql
-- Hobby and creative interest domains

SELECT sender_domain, COUNT(*) as cnt
FROM messages
WHERE sender_domain IN (
  'bandcamp.com','notifications.soundcloud.com','announcements.soundcloud.com',
  'audiotent.com','undergroundmusicacademy.com','timexile.com',
  'magnumphotos.com','riphotocenter.org','iphonephotographyschool.com',
  'news.plugin-alliance.com','play.date','updates.itch.io',
  'bambulab.com','news.criterion.com','charcoalbookclub.com',
  'densediscovery.com','hillelwayne.com','lifestylefilmblog.com',
  'strava.com','email.trainingpeaks.com'
)
GROUP BY sender_domain
ORDER BY cnt DESC;
