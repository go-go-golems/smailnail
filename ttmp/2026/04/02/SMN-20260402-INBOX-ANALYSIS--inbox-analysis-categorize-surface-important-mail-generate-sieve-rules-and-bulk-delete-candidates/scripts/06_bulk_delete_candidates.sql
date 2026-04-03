-- 06_bulk_delete_candidates.sql
-- Messages safe to delete in bulk: high-volume noise with no action value

-- A) Crowdfunding spam (kickstarter-clone spam domains)
SELECT 'crowdfunding-spam' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%kickstarnow.com%'
   OR from_summary LIKE '%kickstarternew.com%'
   OR from_summary LIKE '%kickstartgenius.com%'
   OR from_summary LIKE '%backerhome.com%'
   OR from_summary LIKE '%backerclub.co%'
   OR from_summary LIKE '%huaweiinsider.com%'
   OR from_summary LIKE '%kickstarter-new.net%'
   OR from_summary LIKE '%kickstartrend.com%';

-- B) Firebase / spam phishing (firebaseapp.com noreply = spam)
SELECT 'firebase-spam' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%firebaseapp.com%';

-- C) Twitch stream notifications
SELECT 'twitch-streams' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%no-reply@twitch.tv%'
   AND subject NOT LIKE '%invoice%'
   AND subject NOT LIKE '%receipt%';

-- D) Facebook (reminders, friend suggestions - zero value)
SELECT 'facebook-noise' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%facebookmail.com%';

-- E) Instagram
SELECT 'instagram' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%mail.instagram.com%';

-- F) SoundCloud alerts
SELECT 'soundcloud' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%notifications.soundcloud.com%';

-- G) Zillow instant updates (privaterelay zillow)
SELECT 'zillow' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%zillow%';

-- H) The Tree Center marketing (62 emails!)
SELECT 'thetreecenter-marketing' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%thetreecenter.com%';

-- I) Experian (credit monitoring spam)
SELECT 'experian' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%experian.com%';

-- J) Freelancer.com notifications
SELECT 'freelancer' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%freelancer.com%';

-- K) LinkedIn (notifications, messages - mostly noise)
SELECT 'linkedin' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%linkedin.com%';

-- L) Walgreens
SELECT 'walgreens' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%walgreens.com%';

-- M) CVS (receipts + surveys)
SELECT 'cvs' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%cvs.com%'
   OR from_summary LIKE '%mystore.cvs.com%'
   OR from_summary LIKE '%cvs@express.medallia.com%';

-- N) Manning Publications marketing
SELECT 'manning-marketing' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%manning.com%';

-- O) eBay
SELECT 'ebay' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%ebay.com%';

-- P) ship30for30 / Dickie & Cole
SELECT 'ship30for30' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%ship30for30.com%';

-- Q) Pledge/backer spam
SELECT 'pledge-backer-spam' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%pledgebox.com%';

-- R) Amazon marketing (non-invoice)
SELECT 'amazon-marketing' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%amazon.com%'
  AND subject NOT LIKE '%invoice%'
  AND subject NOT LIKE '%order%'
  AND subject NOT LIKE '%shipped%';

-- S) Domestic/property marketing (Domestika privaterelay, MUJI, TASCHEN)
SELECT 'retail-marketing' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%domestika%'
   OR from_summary LIKE '%muji.us%'
   OR from_summary LIKE '%taschen%'
   OR from_summary LIKE '%mackbooks.co.uk%'
   OR from_summary LIKE '%baronfig.com%'
   OR from_summary LIKE '%news.baronfig.com%';

-- T) iPhone Photography School
SELECT 'iphonephotography-spam' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%iphonephotographyschool.com%';

-- U) Meetup notifications
SELECT 'meetup-notifications' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%email.meetup.com%';

-- GRAND TOTAL bulk-delete estimate
SELECT 'TOTAL_BULK_DELETE' AS bucket, COUNT(*) AS cnt FROM messages
WHERE from_summary LIKE '%kickstarnow.com%'
   OR from_summary LIKE '%kickstarternew.com%'
   OR from_summary LIKE '%kickstartgenius.com%'
   OR from_summary LIKE '%backerhome.com%'
   OR from_summary LIKE '%backerclub.co%'
   OR from_summary LIKE '%huaweiinsider.com%'
   OR from_summary LIKE '%kickstarter-new.net%'
   OR from_summary LIKE '%kickstartrend.com%'
   OR from_summary LIKE '%firebaseapp.com%'
   OR (from_summary LIKE '%no-reply@twitch.tv%' AND subject NOT LIKE '%invoice%')
   OR from_summary LIKE '%facebookmail.com%'
   OR from_summary LIKE '%mail.instagram.com%'
   OR from_summary LIKE '%notifications.soundcloud.com%'
   OR from_summary LIKE '%zillow%'
   OR from_summary LIKE '%thetreecenter.com%'
   OR from_summary LIKE '%experian.com%'
   OR from_summary LIKE '%freelancer.com%'
   OR from_summary LIKE '%linkedin.com%'
   OR from_summary LIKE '%walgreens.com%'
   OR (from_summary LIKE '%cvs.com%' OR from_summary LIKE '%mystore.cvs.com%' OR from_summary LIKE '%medallia.com%')
   OR from_summary LIKE '%manning.com%'
   OR from_summary LIKE '%ebay.com%'
   OR from_summary LIKE '%ship30for30.com%'
   OR from_summary LIKE '%pledgebox.com%'
   OR (from_summary LIKE '%domestika%' OR from_summary LIKE '%muji.us%' OR from_summary LIKE '%taschen%' OR from_summary LIKE '%mackbooks%' OR from_summary LIKE '%baronfig%')
   OR from_summary LIKE '%iphonephotographyschool.com%'
   OR from_summary LIKE '%email.meetup.com%';
