#!/usr/bin/env bash
# 02-categorize-newsletter-senders.sh — Tag newsletter senders by subcategory
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

echo "=== Tech newsletters ==="
# Substack tech writers
for email in \
  "garymarcus@substack.com" \
  "pragmaticengineer@substack.com" \
  "pragmaticengineer+deepdives@substack.com" \
  "simonw@substack.com" \
  "swyx@substack.com" \
  "swyx+ainews@substack.com" \
  "escapingflatland@substack.com" \
  "howthingswork@substack.com" \
  "thursdai@substack.com" \
  "frontierai@substack.com" \
  "superintelligencenews@substack.com" \
  "robotic@substack.com" \
  "countercraft@substack.com" \
  "read@substack.com" \
  "no-reply@substack.com"
do
  annotate_sender "$email" "newsletter/tech" "Tech newsletter (Substack)"
done

# Non-substack tech
annotate_sender "hello@every.to" "newsletter/tech" "Every.to tech/business writing (426 msgs)"
annotate_sender "hello@readwise.io" "newsletter/tech" "Readwise digest (396 msgs)"
annotate_sender "turingpost@mail.beehiiv.com" "newsletter/tech" "Turing Post AI newsletter (105 msgs)"
annotate_sender "alphasignal@alphasignal.ai" "newsletter/tech" "Alpha Signal AI newsletter (66 msgs)"
annotate_sender "emil@iphonephotographyschool.com" "newsletter/tech" "iPhone Photography School (272 msgs)"
annotate_sender "kai@threedotslabs.com" "newsletter/tech" "Three Dots Labs tech newsletter"
annotate_sender "hello@densediscovery.com" "newsletter/tech" "Dense Discovery design/tech newsletter (82 msgs)"
annotate_sender "hillelwayne@hillelwayne.com" "newsletter/tech" "Hillel Wayne formal methods/testing newsletter (51 msgs)"

echo ""
echo "=== Culture newsletters ==="
annotate_sender "maxread@substack.com" "newsletter/culture" "Max Read culture/internet newsletter (81 msgs)"
annotate_sender "bloodinthemachine@substack.com" "newsletter/culture" "Blood in the Machine labor/tech culture (85 msgs)"
annotate_sender "drdevonprice@substack.com" "newsletter/culture" "Devon Price essays (66 msgs)"
annotate_sender "theleverage@substack.com" "newsletter/culture" "The Leverage newsletter (88 msgs)"
annotate_sender "eleanorkonik@substack.com" "newsletter/culture" "Eleanor Konik knowledge mgmt/history (54 msgs)"
annotate_sender "lindac@substack.com" "newsletter/culture" "Linda C newsletter (46 msgs)"
annotate_sender "hello@charcoalbookclub.com" "newsletter/culture" "Charcoal Book Club (76 msgs)"
annotate_sender "emails@emails.theintercept.com" "newsletter/culture" "The Intercept journalism (61 msgs)"

echo ""
echo "=== Creative newsletters ==="
annotate_sender "noreply@bandcamp.com" "newsletter/creative" "Bandcamp new releases/recommendations (96 msgs)"
annotate_sender "announcements@announcements.soundcloud.com" "newsletter/creative" "SoundCloud announcements (70 msgs)"
annotate_sender "hello@audiotent.com" "newsletter/creative" "Audio Tent music production (88 msgs)"
annotate_sender "tim@timexile.com" "newsletter/creative" "Tim Exile music tech (57 msgs)"
annotate_sender "hello@undergroundmusicacademy.com" "newsletter/creative" "Underground Music Academy (50 msgs)"
annotate_sender "hello@lifestylefilmblog.com" "newsletter/creative" "Lifestyle Film Blog photography (55 msgs)"
annotate_sender "criterion@news.criterion.com" "newsletter/creative" "Criterion Collection film (60 msgs)"
annotate_sender "postmaster@updates.itch.io" "newsletter/creative" "itch.io indie games updates (94 msgs)"
annotate_sender "info@ecamm.com" "newsletter/creative" "Ecamm streaming/video (52 msgs)"

echo ""
echo "=== Done. Annotated $COUNT newsletter senders ==="
