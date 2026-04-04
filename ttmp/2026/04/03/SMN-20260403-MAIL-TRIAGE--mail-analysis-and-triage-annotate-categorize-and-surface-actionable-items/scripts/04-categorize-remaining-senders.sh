#!/usr/bin/env bash
# 04-categorize-remaining-senders.sh — Sweep up remaining high-volume unannotated senders
set -euo pipefail
DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"
COUNT=0

annotate_sender() {
  local email="$1" tag="$2" note="$3"
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations WHERE target_type='sender' AND target_id='$email' AND tag='$tag';")
  if [ "$existing" -gt 0 ]; then return; fi
  smailnail annotate annotation add --sqlite-path "$DB" \
    --target-type sender --target-id "$email" --tag "$tag" --note "$note" \
    --source-kind heuristic --source-label mail-triage-v1 --created-by pi-agent \
    --output json > /dev/null 2>&1
  echo "OK: $email -> $tag"
  COUNT=$((COUNT + 1))
}

echo "=== Apple Private Relay senders ==="
# These are services accessed via Apple Sign-In
annotate_sender "rover_at_e_rover_com_mqbtnkv9gk_3ff7b9ab@privaterelay.appleid.com" "services" "Rover pet services via Apple relay (403 msgs)"
annotate_sender "no-reply_at_doordash_com_q28t5qzvsm_ec003cb1@privaterelay.appleid.com" "services" "DoorDash via Apple relay (108 msgs)"
annotate_sender "domestika_at_news_domestika_org_jmfs8qcnxe_5a3b88b9@privaterelay.appleid.com" "hobby" "Domestika online courses via Apple relay (90 msgs)"
annotate_sender "news_at_m_elements_envato_com_g7qdzddct7_43dfc8ff@privaterelay.appleid.com" "noise/marketing" "Envato Elements marketing via Apple relay (55 msgs)"

echo ""
echo "=== More hobby/creative senders ==="
annotate_sender "info@audiotent.com" "hobby" "Audio Tent sample packs/production (88 msgs)"
annotate_sender "exile@timexile.com" "hobby" "Tim Exile music tech creator (57 msgs)"
annotate_sender "thomas@lifestylefilmblog.com" "hobby" "Lifestyle Film Blog photography (55 msgs)"

echo ""
echo "=== More community senders ==="
annotate_sender "gallery@riphotocenter.org" "community" "RI Photo Center gallery events (88 msgs)"
annotate_sender "david@as220.org" "community" "AS220 community events (59 msgs)"
annotate_sender "mail@provath.org" "community" "ProVath organization (66 msgs)"

echo ""
echo "=== More services ==="
annotate_sender "buero@gewerbehof-karlsruhe.de" "services" "Gewerbehof Karlsruhe office (81 msgs)"
annotate_sender "support@digitalocean.com" "services" "DigitalOcean support (72 msgs)"
annotate_sender "support@e.usa.experian.com" "financial" "Experian credit monitoring (68 msgs)"
annotate_sender "email@e.affirm.com" "financial" "Affirm payment notifications (82 msgs)"
annotate_sender "customerservice@sales.sub-shop.com" "services" "Sub-Shop customer service (79 msgs)"
annotate_sender "no-reply@latch.com" "services" "Latch smart lock/building already done"

echo ""
echo "=== More tech/work ==="
annotate_sender "developer@insideapple.apple.com" "newsletter/tech" "Apple Developer news (80 msgs)"
annotate_sender "applied-llms@courses.maven.com" "newsletter/tech" "Applied LLMs Maven course (68 msgs)"
annotate_sender "wandb@mail.wandb.ai" "newsletter/tech" "Weights & Biases ML newsletter (57 msgs)"
annotate_sender "robert.laszczak@threedotslabs.com" "newsletter/tech" "Three Dots Labs Go newsletter (56 msgs)"
annotate_sender "news@alphasignal.ai" "newsletter/tech" "Alpha Signal AI newsletter (66 msgs)"

echo ""
echo "=== More social/noise ==="
annotate_sender "info@twitter.com" "noise/social-notif" "Twitter/X notifications (58 msgs)"
annotate_sender "groups-noreply@linkedin.com" "noise/social-notif" "LinkedIn group notifications (58 msgs)"
annotate_sender "noreply@skool.com" "noise/social-notif" "Skool community notifications (65 msgs)"
annotate_sender "info@announcements.soundcloud.com" "noise/social-notif" "SoundCloud announcements (70 msgs)"

echo ""
echo "=== More newsletters ==="
annotate_sender "info@charcoalbookclub.com" "newsletter/culture" "Charcoal Book Club (76 msgs)"
annotate_sender "newsletter@news.criterion.com" "newsletter/creative" "Criterion Collection (60 msgs)"
annotate_sender "no-reply@emails.theintercept.com" "newsletter/culture" "The Intercept journalism (60 msgs)"
annotate_sender "newsletter@mackbooks.co.uk" "newsletter/creative" "Mack Books photography (59 msgs)"
annotate_sender "newsletter@backerclub.co" "noise/marketing" "Backer Club crowdfunding marketing (59 msgs)"
annotate_sender "info@email.trainingpeaks.com" "hobby" "TrainingPeaks fitness newsletter (57 msgs)"

echo ""
echo "=== More remaining senders from 30-55 msg range ==="

# LinkedIn
annotate_sender "messages-noreply@linkedin.com" "noise/social-notif" "LinkedIn message notifications"
annotate_sender "invitations@linkedin.com" "noise/social-notif" "LinkedIn connection requests"
annotate_sender "jobs-noreply@linkedin.com" "noise/social-notif" "LinkedIn job alerts"
annotate_sender "notifications-noreply@linkedin.com" "noise/social-notif" "LinkedIn notifications"

# More Apple
annotate_sender "noreply@insideapple.apple.com" "newsletter/tech" "Apple Inside developer news"
annotate_sender "appleid@id.apple.com" "services" "Apple ID notifications"

# More Substack
annotate_sender "ainews@substack.com" "newsletter/tech" "AI News substack"
annotate_sender "simonsarris@substack.com" "newsletter/culture" "Simon Sarris essays"

# E-commerce/marketing more
annotate_sender "noreply@patreon.com" "services" "Patreon creator platform notifications (43 msgs)"
annotate_sender "promo@email.meetup.com" "noise/marketing" "Meetup promotional emails"
annotate_sender "noreply@m.mail.coursera.org" "noise/marketing" "Coursera marketing emails (41 msgs)"
annotate_sender "noreply@e.stripe.com" "services" "Stripe payment notifications (62 msgs)"
annotate_sender "noreply@calendar.luma-mail.com" "services" "Luma calendar event invites (59 msgs)"
annotate_sender "hello@davemillercpa.com" "services" "Dave Miller CPA accounting (49 msgs)"
annotate_sender "llm@llamaindex.ai" "newsletter/tech" "LlamaIndex AI newsletter (40 msgs)"
annotate_sender "noreply@marketing.descript.com" "noise/marketing" "Descript marketing (32 msgs)"

echo ""
echo "=== Remaining large Apple Private Relay senders ==="
sqlite3 "$DB" "
SELECT sender_email, COUNT(*) as cnt 
FROM messages 
WHERE sender_domain='privaterelay.appleid.com' AND sender_email != ''
  AND NOT EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=messages.sender_email)
GROUP BY sender_email 
HAVING cnt >= 10
ORDER BY cnt DESC;" | while IFS='|' read -r email cnt; do
  annotate_sender "$email" "services" "Apple Private Relay service sender ($cnt msgs)"
done

echo ""
echo "=== Done. Total new annotations: $COUNT ==="
