#!/usr/bin/env bash
# 03-categorize-personal-work-services.sh — Tag personal, work, community, service, hobby, financial senders
set -euo pipefail
DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"
COUNT=0

annotate_sender() {
  local email="$1" tag="$2" note="$3"
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations WHERE target_type='sender' AND target_id='$email' AND tag='$tag';")
  if [ "$existing" -gt 0 ]; then
    echo "SKIP: $email -> $tag"
    return
  fi
  smailnail annotate annotation add --sqlite-path "$DB" \
    --target-type sender --target-id "$email" --tag "$tag" --note "$note" \
    --source-kind heuristic --source-label mail-triage-v1 --created-by pi-agent \
    --output json > /dev/null 2>&1
  echo "OK: $email -> $tag"
  COUNT=$((COUNT + 1))
}

echo "=== Personal correspondents ==="
# Top personal senders identified from gmail/icloud/hotmail/yahoo analysis
annotate_sender "ginny.white@gmail.com" "personal" "Active personal correspondent (42 msgs)"
annotate_sender "perryholser@gmail.com" "personal" "Active personal correspondent (34 msgs)"
annotate_sender "robokid13@gmail.com" "personal" "Active personal correspondent, SOITS Alumni thread (33 msgs)"
annotate_sender "erinmperfect@gmail.com" "personal" "Active personal correspondent (33 msgs)"
annotate_sender "gleitman.sam@gmail.com" "personal" "Active personal correspondent, SOITS Alumni (27 msgs)"
annotate_sender "fabianfaedrich1@gmail.com" "personal" "Active personal correspondent (21 msgs)"
annotate_sender "kryzaklaw@yahoo.com" "personal" "Personal correspondent (19 msgs)"
annotate_sender "hobbydavid@yahoo.com" "personal" "Personal correspondent (12 msgs)"
annotate_sender "bessetodendahl@icloud.com" "personal" "Personal correspondent (7 msgs)"
annotate_sender "newoldtraditions@gmail.com" "personal" "Personal correspondent, Zettelkasten (6 msgs)"
annotate_sender "millerdevel@gmail.com" "personal" "Personal correspondent (6 msgs)"
annotate_sender "koshy44@gmail.com" "personal" "Personal correspondent (6 msgs)"
annotate_sender "brendan.leonard111@gmail.com" "personal" "Personal correspondent, DPW/community (6 msgs)"
annotate_sender "jonwalkerphoto@gmail.com" "personal" "Personal correspondent (5 msgs)"
annotate_sender "goldhamstermoegen@icloud.com" "personal" "Personal correspondent (5 msgs)"
annotate_sender "fahree@gmail.com" "personal" "Personal correspondent, method dispatch discussion (5 msgs)"
annotate_sender "avinashsajjanshetty@gmail.com" "personal" "Personal correspondent (5 msgs)"
annotate_sender "3.4.5.6.7.8.9.zehn.elf@gmail.com" "personal" "Personal correspondent (5 msgs)"
annotate_sender "hans.huebner@gmail.com" "personal" "Personal correspondent (4 msgs)"
annotate_sender "adriani.botez@gmail.com" "personal" "Personal correspondent (4 msgs)"
annotate_sender "achambers.home@gmail.com" "personal" "Personal correspondent, method dispatch discussion (recent)"
annotate_sender "besset.stadler@gmail.com" "personal" "Personal correspondent, Ascension La Forclaz (recent)"
annotate_sender "lepetitg@gmail.com" "personal" "Personal correspondent, community (4 msgs)"
annotate_sender "learncamerarepair@gmail.com" "personal" "Personal correspondent (4 msgs)"
annotate_sender "jonathanmarkfisher@gmail.com" "personal" "Personal correspondent, meeting link (recent)"
annotate_sender "mikael.francoeur@hotmail.com" "personal" "Personal correspondent (4 msgs)"
annotate_sender "stepanart7@gmail.com" "personal" "Personal correspondent, concert invite (3 msgs)"

echo ""
echo "=== Work senders ==="
annotate_sender "notifier@mail.rollbar.com" "work" "Rollbar error monitoring (400 msgs)"
annotate_sender "noreply@zulip.com" "work" "Zulip chat notifications — Recurse Center (74 msgs)"
annotate_sender "feedback@slack.com" "work" "Slack notifications (66 msgs)"
annotate_sender "notification@slack.com" "work" "Slack notifications (49 msgs)"
annotate_sender "no-reply@slack.com" "work" "Slack account notifications (11 msgs)"
annotate_sender "no-reply@carta.com" "work" "Carta equity/HR (7 msgs)"
annotate_sender "support@chromatic.com" "work" "Chromatic (Storybook CI) (4 msgs)"

echo ""
echo "=== Community senders ==="
annotate_sender "intern@lists.entropia.de" "community" "Entropia hackerspace mailing list (420 msgs)"
annotate_sender "info@as220.org" "community" "AS220 arts community (59 msgs)"
annotate_sender "info@riphotocenter.org" "community" "RI Photo Center community (94 msgs)"
annotate_sender "dreeves@beeminder.com" "community" "Beeminder monthly beemail (recent)"
annotate_sender "mwillia8@bates.edu" "community" "Bates community, film/art screenings (recent)"

echo ""
echo "=== Financial senders ==="
annotate_sender "support@e.affirm.com" "financial" "Affirm payment notifications (82 msgs)"
annotate_sender "no-reply@post.applecard.apple" "financial" "Apple Card statements (48 msgs)"

echo ""
echo "=== Services senders ==="
annotate_sender "no-reply@accounts.google.com" "services" "Google account notifications"
annotate_sender "noreply@uber.com" "services" "Uber ride/delivery notifications"
annotate_sender "noreply@chewy.com" "services" "Chewy pet supplies (39 msgs)"
annotate_sender "ebay@ebay.com" "services" "eBay notifications (86 msgs)"
annotate_sender "noreply@airbnb.com" "services" "Airbnb notifications (60 msgs)"
annotate_sender "no-reply@latch.com" "services" "Latch smart lock/building (47 msgs)"
annotate_sender "hello@modernmsg.com" "services" "Modern Message property management (38 msgs)"
annotate_sender "no-reply@digitalocean.com" "services" "DigitalOcean hosting (72 msgs)"
annotate_sender "vorstand@gewerbehof-karlsruhe.de" "services" "Gewerbehof Karlsruhe (94 msgs)"
annotate_sender "noreply@prosperhealth.io" "services" "Prosper Health notifications (49 msgs)"

echo ""
echo "=== Hobby senders ==="
annotate_sender "noreply@bambulab.com" "hobby" "Bambu Lab 3D printing (43 msgs)"
annotate_sender "hello@play.date" "hobby" "Playdate handheld gaming (41 msgs)"
annotate_sender "noreply@strava.com" "hobby" "Strava fitness tracking (45 msgs)"
annotate_sender "noreply@email.trainingpeaks.com" "hobby" "TrainingPeaks fitness (57 msgs)"
annotate_sender "noreply@todoist.com" "hobby" "Todoist productivity (43 msgs)"
annotate_sender "hello@svslearn.com" "hobby" "SVS Learn art education (39 msgs)"

echo ""
echo "=== More noise/marketing domains (bulk) ==="
# Additional marketing senders found in exploration
for domain in visionreviewed.com activelife.za.com checkspot.za.com chiefdull.my fulltechstore.my hometechgear.my realtechdeals.my striploot.za.com termbreed.sa.com xtremeninja.com; do
  sqlite3 "$DB" "SELECT sender_email FROM messages WHERE sender_domain='$domain' AND sender_email != '' GROUP BY sender_email;" | while read -r email; do
    annotate_sender "$email" "noise/spam" "Sender at spam domain $domain"
  done
done

echo ""
echo "=== Done. Total new annotations: $COUNT ==="
