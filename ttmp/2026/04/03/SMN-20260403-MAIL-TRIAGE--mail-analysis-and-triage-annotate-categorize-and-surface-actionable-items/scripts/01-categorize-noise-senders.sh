#!/usr/bin/env bash
# 01-categorize-noise-senders.sh — Tag noise senders: CI, marketing, spam, social, transactional
set -euo pipefail
DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"
SN="smailnail annotate annotation add --sqlite-path $DB --source-kind heuristic --source-label mail-triage-v1 --created-by pi-agent"
COUNT=0

annotate_sender() {
  local email="$1" tag="$2" note="$3"
  # Skip if already annotated with this tag
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations WHERE target_type='sender' AND target_id='$email' AND tag='$tag';")
  if [ "$existing" -gt 0 ]; then
    echo "SKIP (exists): $email -> $tag"
    return
  fi
  $SN --target-type sender --target-id "$email" --tag "$tag" --note "$note" --output json > /dev/null 2>&1
  echo "OK: $email -> $tag"
  COUNT=$((COUNT + 1))
}

annotate_domain() {
  local domain="$1" tag="$2" note="$3"
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations WHERE target_type='domain' AND target_id='$domain' AND tag='$tag';")
  if [ "$existing" -gt 0 ]; then
    echo "SKIP (exists): domain:$domain -> $tag"
    return
  fi
  $SN --target-type domain --target-id "$domain" --tag "$tag" --note "$note" --output json > /dev/null 2>&1
  echo "OK: domain:$domain -> $tag"
  COUNT=$((COUNT + 1))
}

echo "=== Phase 1a: Noise/CI senders ==="
# notifications@github.com already annotated in test run
annotate_sender "noreply@github.com" "noise/ci" "GitHub noreply (234 msgs) — automated notifications"

echo ""
echo "=== Phase 1b: Noise/spam domains ==="
for domain in purchasingreviews.com costsoldier.com voip-prices.com small-business-search.com smallbusinesspurchasing.com mail1.bostonmarketinggroup.com; do
  annotate_domain "$domain" "noise/spam" "Spam/bulk marketing domain"
  # Also annotate individual senders
  sqlite3 "$DB" "SELECT sender_email FROM messages WHERE sender_domain='$domain' AND sender_email != '' GROUP BY sender_email;" | while read -r email; do
    annotate_sender "$email" "noise/spam" "Sender at spam domain $domain"
  done
done

echo ""
echo "=== Phase 1c: Noise/marketing senders ==="
declare -A MARKETING_SENDERS=(
  ["sales@thetreecenter.com"]="The Tree Center marketing (865 msgs)"
  ["mkt@manning.com"]="Manning Publications marketing (170 msgs)"
  ["hello@ship30for30.com"]="Ship30for30 marketing (157 msgs)"
  ["walgreens@eml.walgreens.com"]="Walgreens marketing/coupons (155 msgs)"
  ["newsletter@news.plugin-alliance.com"]="Plugin Alliance marketing (139 msgs)"
  ["us@fullstack.io"]="Fullstack.io marketing (116 msgs)"
  ["noreply@notifications.freelancer.com"]="Freelancer.com job spam (320 msgs)"
  ["magnum@magnumphotos.com"]="Magnum Photos marketing (103 msgs)"
  ["extracare@your.cvs.com"]="CVS ExtraCare marketing (207 msgs)"
  ["email@e.godaddy.com"]="GoDaddy marketing"
  ["hello@mackbooks.co.uk"]="Mack Books marketing (59 msgs)"
  ["info@backerclub.co"]="Backer Club marketing (59 msgs)"
)
for email in "${!MARKETING_SENDERS[@]}"; do
  annotate_sender "$email" "noise/marketing" "${MARKETING_SENDERS[$email]}"
done

echo ""
echo "=== Phase 1d: Noise/social-notif senders ==="
declare -A SOCIAL_SENDERS=(
  ["no-reply@twitch.tv"]="Twitch notifications (419 msgs)"
  ["no-reply@is.email.nextdoor.com"]="Nextdoor notifications (217 msgs)"
  ["alerts@notifications.soundcloud.com"]="SoundCloud alerts (99 msgs)"
  ["noreply@facebookmail.com"]="Facebook notifications"
  ["notification@facebookmail.com"]="Facebook notifications"
  ["info@email.meetup.com"]="Meetup event notifications (697 msgs)"
  ["no-reply@rs.email.nextdoor.com"]="Nextdoor digest (87 msgs)"
)
for email in "${!SOCIAL_SENDERS[@]}"; do
  annotate_sender "$email" "noise/social-notif" "${SOCIAL_SENDERS[$email]}"
done

echo ""
echo "=== Phase 1e: Noise/transactional senders ==="
declare -A TRANSACT_SENDERS=(
  ["shipment-tracking@amazon.com"]="Amazon shipment tracking (307 msgs)"
  ["auto-confirm@amazon.com"]="Amazon order confirmations (131 msgs)"
  ["order-update@amazon.com"]="Amazon order updates (261 msgs)"
  ["service@paypal.com"]="PayPal transaction notifications (191 msgs)"
  ["no_reply@email.apple.com"]="Apple receipts/notifications (191 msgs)"
  ["buchung@karlsruhe.stadtmobil.de"]="Stadtmobil booking confirmations (233 msgs)"
  ["venmo@venmo.com"]="Venmo transaction notifications"
  ["onlinebanking@ealerts.bankofamerica.com"]="Bank of America alerts (101 msgs)"
  ["halsteadprovidence@bozzuto.com"]="Bozzuto property management (181 msgs)"
)
for email in "${!TRANSACT_SENDERS[@]}"; do
  annotate_sender "$email" "noise/transactional" "${TRANSACT_SENDERS[$email]}"
done

echo ""
echo "=== Done. Annotated $COUNT senders/domains ==="
