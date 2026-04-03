#!/usr/bin/env bash
# 08-categorize-high-importance.sh — Tag high-importance senders: tax/CPA, lawyer, equity, housing, work admin
set -euo pipefail
DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"
COUNT=0

a() {
  local email="$1" tag="$2" note="$3"
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations WHERE target_type='sender' AND target_id='$email' AND tag='$tag';")
  if [ "$existing" -gt 0 ]; then echo "SKIP: $email"; return; fi
  smailnail annotate annotation add --sqlite-path "$DB" \
    --target-type sender --target-id "$email" --tag "$tag" --note "$note" \
    --source-kind heuristic --source-label mail-triage-v1 --created-by pi-agent \
    --output json > /dev/null 2>&1
  echo "OK: $email -> $tag"
  COUNT=$((COUNT + 1))
}

echo "=== TAX / CPA ==="
a "dave@davemillercpa.com" "important/tax" "CPA - David Miller, primary tax preparer"
a "liam@davemillercpa.com" "important/tax" "CPA office - Liam at David Miller CPA, tax questions"
a "office@davemillercpa.com" "important/tax" "CPA office - billing, e-filing, tax prep"
a "info@davemillercpa.com" "important/tax" "CPA office - investment commentary"
a "quickbooks@notification.intuit.com" "important/tax" "QuickBooks/Intuit invoices from CPA"

echo ""
echo "=== LAWYER / LEGAL ==="
a "kryzaklaw@yahoo.com" "important/legal" "Lawyer - Kryzak, separation agreement (19 msgs)"
a "ellen@wrightfamilylawgroup.com" "important/legal" "Wright Family Law Group - Ellen, workshops & support"
a "elaine@wrightfamilylawgroup.com" "important/legal" "Wright Family Law Group - Elaine, retainer/billing"
a "emily@wrightfamilylawgroup.com" "important/legal" "Wright Family Law Group - Emily, check-ins"

echo ""
echo "=== EQUITY / STOCK OPTIONS ==="
a "tripp@equitybee.com" "important/equity" "EquityBee - Formlabs stock options inquiry"
a "supply@equityzen.com" "important/equity" "EquityZen - Formlabs holdings confirmation requests"
a "support@equityzen.com" "important/equity" "EquityZen - private market pulse, policy updates"
a "noreply@mail.hiive.com" "important/equity" "Hiive - startup equity selling platform"
a "nathan.eraut@mail.hiive.com" "important/equity" "Hiive - personal outreach"
a "investors@equitybee.com" "important/equity" "EquityBee - investing platform"

echo ""
echo "=== WORK / EMPLOYER ADMIN ==="
a "donotreply@admoove.com" "important/work-admin" "Admoove - bl0rg end-of-year compensation"
a "no-reply@carta.com" "important/work-admin" "Carta - equity/cap table management"

echo ""
echo "=== HOUSING / RENT ==="
a "notifications@alerts.biltrewards.com" "important/housing" "Bilt Rewards - rent autopay processing"
a "notifications@members.biltrewards.com" "important/housing" "Bilt Rewards - rent day benefits"
a "no-reply@notifications.biltrewards.com" "important/housing" "Bilt Rewards - account verification"
a "halsteadprovidence@bozzuto.com" "important/housing" "Bozzuto/Halstead Providence - property management (181 msgs)"

echo ""
echo "=== HEALTH ==="
a "admin@revivetherapeuticservices.com" "important/health" "Revive Therapeutic Services - medication management"
a "noreply@prosperhealth.io" "important/health" "Prosper Health - health notifications"
a "support@phoenixrisingcenters.org" "important/health" "Phoenix Rising Centers - onboarding"

echo ""
echo "=== CLASS ACTION / LEGAL NOTICES ==="
a "lucchesesoto@e.emailksa.com" "important/legal" "Class action settlement legal notice"

echo ""
echo "=== Done. Total new: $COUNT ==="
