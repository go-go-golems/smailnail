-- 03_categorize.sql
-- Assign a category to every message based on sender/subject heuristics
-- Categories: github, newsletter, social, commercial, security, automated, mailing-list, personal, other

WITH categorized AS (
  SELECT
    id, from_summary, subject, sent_date, size_bytes,
    CASE
      -- GitHub notifications
      WHEN from_summary LIKE '%notifications@github.com%'
        OR from_summary LIKE '%noreply@github.com%'
        THEN 'github'
      -- Mailing lists (Mailman / riseup / entropia / etc.)
      WHEN from_summary LIKE '%lists.riseup.net%'
        OR from_summary LIKE '%lists.entropia.de%'
        OR from_summary LIKE '%@lists.%'
        OR subject LIKE '[%]%'
        THEN 'mailing-list'
      -- Newsletters / Substack / Beehiiv
      WHEN from_summary LIKE '%substack.com%'
        OR from_summary LIKE '%@mg1.substack.com%'
        OR from_summary LIKE '%mail.beehiiv.com%'
        OR from_summary LIKE '%every.to%'
        OR from_summary LIKE '%ship30for30.com%'
        OR from_summary LIKE '%readwise.io%'
        THEN 'newsletter'
      -- Social / LinkedIn / Facebook / SoundCloud / Twitch
      WHEN from_summary LIKE '%linkedin.com%'
        OR from_summary LIKE '%facebookmail.com%'
        OR from_summary LIKE '%notifications.soundcloud.com%'
        OR from_summary LIKE '%twitch.tv%'
        THEN 'social'
      -- Security / account alerts
      WHEN from_summary LIKE '%accountprotection.microsoft.com%'
        OR from_summary LIKE '%accounts.google.com%'
        OR subject LIKE '%security%'
        OR subject LIKE '%verify%'
        OR subject LIKE '%sign-in%'
        OR subject LIKE '%password%'
        THEN 'security'
      -- Financial / shopping
      WHEN from_summary LIKE '%paypal.com%'
        OR from_summary LIKE '%bankofamerica.com%'
        OR from_summary LIKE '%ebay.com%'
        OR from_summary LIKE '%e.affirm.com%'
        OR from_summary LIKE '%experian.com%'
        OR from_summary LIKE '%e.usa.experian.com%'
        OR from_summary LIKE '%walgreens.com%'
        OR from_summary LIKE '%amazon.com%'
        OR from_summary LIKE '%cvs.com%'
        THEN 'financial-shopping'
      -- Commercial / marketing / crowdfunding
      WHEN from_summary LIKE '%manning.com%'
        OR from_summary LIKE '%thetreecenter.com%'
        OR from_summary LIKE '%pledgebox.com%'
        OR from_summary LIKE '%kickstarnow.com%'
        OR from_summary LIKE '%kickstarternew.com%'
        OR from_summary LIKE '%kickstartgenius.com%'
        OR from_summary LIKE '%backerhome.com%'
        OR from_summary LIKE '%backerclub.co%'
        OR from_summary LIKE '%muji.us%'
        OR from_summary LIKE '%taschen%'
        OR from_summary LIKE '%domestika%'
        OR from_summary LIKE '%iphonephotographyschool.com%'
        OR from_summary LIKE '%freelancer.com%'
        OR from_summary LIKE '%replit.com%'
        OR from_summary LIKE '%augmentcode.com%'
        OR from_summary LIKE '%ecamm.com%'
        OR from_summary LIKE '%itch.io%'
        OR from_summary LIKE '%firecrawl.dev%'
        OR from_summary LIKE '%envato%'
        OR from_summary LIKE '%privaterelay.appleid.com%'
        OR from_summary LIKE '%sales.sub-shop.com%'
        THEN 'commercial'
      -- Automated / system / monitoring
      WHEN from_summary LIKE '%rollbar.com%'
        OR from_summary LIKE '%no-reply@%'
        OR from_summary LIKE '%noreply@%'
        OR from_summary LIKE '%[bot]%'
        OR from_summary LIKE '%notification%'
        OR from_summary LIKE '%alert%'
        OR from_summary LIKE '%automated%'
        OR from_summary LIKE '%meetup.com%'
        OR from_summary LIKE '%bozzuto.com%'
        OR from_summary LIKE '%patient-message.com%'
        OR from_summary LIKE '%zulip.com%'
        THEN 'automated'
      -- Local/personal/gmail
      WHEN from_summary LIKE '%gmail.com%'
        OR from_summary LIKE '%@bl0rg.net%'
        THEN 'personal'
      ELSE 'other'
    END AS category
  FROM messages
)
SELECT
  category,
  COUNT(*) AS msg_count,
  ROUND(100.0 * COUNT(*) / (SELECT COUNT(*) FROM messages), 1) AS pct,
  ROUND(SUM(size_bytes)/1048576.0, 2) AS total_mb
FROM categorized
GROUP BY category
ORDER BY msg_count DESC;
