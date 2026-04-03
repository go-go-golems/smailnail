---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../../../tmp/smailnail-last-month.sqlite
      Note: Source SQLite database analyzed
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/03_categorize.sql
      Note: Category breakdown driving report numbers
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/06_bulk_delete_candidates.sql
      Note: Bulk-delete counts in report
    - Path: smailnail/ttmp/2026/04/02/SMN-20260402-INBOX-ANALYSIS--inbox-analysis-categorize-surface-important-mail-generate-sieve-rules-and-bulk-delete-candidates/scripts/08_sieve_rules.sieve
      Note: Sieve file referenced by report
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Inbox Analysis Report

**Account:** `mail-bl0rg-net-993-manuel` (bl0rg.net)  
**Period:** 2026-03-02 → 2026-04-02 (one month)  
**Total messages:** 2,894  
**Analysis date:** 2026-04-02

---

## Executive Summary

Your INBOX is **overwhelmingly automated noise**: 97% of email is newsletters, GitHub CI alerts, commercial marketing, social pings, or financial auto-notifications. Only ~8 emails (0.3%) are personal human-written messages. **889 messages (31%) can be safely deleted right now** with no information loss. A further ~1,600 messages can be auto-filed to folders, reducing your inbox to roughly 150–200 genuinely relevant items per month.

Three items need immediate attention:
- ⚠️ **Hetzner Final Payment Warning** — server K1264826925 may be blocked
- ⚠️ **Google Critical Security Alert** (×3, March 30) — wesen3000@googlemail.com
- ⚠️ **Microsoft Unusual Login Activity** (×6, March 13) — may indicate credential exposure

---

## Category Breakdown

| Category | Count | % | MB |
|----------|------:|---:|---:|
| GitHub notifications | 771 | 26.6% | 25.6 |
| Commercial/marketing | 505 | 17.4% | 38.5 |
| Uncategorized (see below) | 454 | 15.7% | 33.5 |
| Newsletters | 393 | 13.6% | 38.3 |
| Automated/system | 283 | 9.8% | 12.8 |
| Social | 179 | 6.2% | 6.0 |
| Financial/shopping | 149 | 5.1% | 11.3 |
| Mailing lists | 116 | 4.0% | 7.2 |
| Security | 36 | 1.2% | 0.5 |
| **Personal (human)** | **8** | **0.3%** | 2.6 |

The "uncategorized" 454 messages break down into: photography communities, art/culture venues, AI/ML tool marketing, proxy/scraping services, local community orgs, and personal acquaintances.

---

## GitHub: The #1 Noise Source (771 messages, 27%)

| Sub-type | Count |
|----------|------:|
| **CI failures** | **652** (84%) |
| PR/issue discussions | 97 |
| GitHub system (billing, etc.) | 22 |

CI failures dominate entirely. The Codex bot (`chatgpt-codex-connector[bot]`) alone generated **96 notifications** in one month. Top repos:

| Repo | Count |
|------|------:|
| go-go-golems/\* (other) | 382 |
| wesen/\* (other) | 119 |
| go-go-golems/geppetto | 72 |
| go-go-golems/pinocchio | 71 |
| wesen/goldeneaglecoin.com | 31 |
| go-go-golems/bobatea | 23 |
| go-go-golems/go-go-os-frontend | 22 |
| wesen/temporal-relationships | 17 |
| go-go-golems/codex-sessions | 11 |

**Recommendation:** File all CI failures to `github/ci`, PR discussions to `github/prs`. Review `goldeneaglecoin.com` CI failures (31 messages — prod deploy failures).

---

## Newsletters (393 messages, 14%)

Substack dominates. Notable subscriptions:

| Newsletter | Count | Keep? |
|-----------|------:|-------|
| Substack (generic wrapper) | 52 | Route to folder |
| AINews (swyx) | 22 | ✅ Keep |
| Gary Marcus on AI | 16 | ✅ Keep |
| Linda Caroll / Hello Writer | 12 | ❓ Review |
| The Pragmatic Engineer | 10+4 | ✅ Keep |
| Hamilton Nolan / How Things Work | 10 | ✅ Keep |
| Latent.Space (swyx) | 9 | ✅ Keep |
| Max Read | 8 | ✅ Keep |
| Slavoj Žižek / Politics | 7 | ✅ Keep |
| Interconnects (Nathan Lambert) | 7 | ✅ Keep |
| Counter Craft | 6 | ❓ Review |
| Readwise | 24 | Route to folder |
| Every.to | 33 | Route to folder |

**Recommendation:** All newsletters → `newsletters/substack` (or sub-folders). Unsubscribe from Hello Writer and Counter Craft if not reading.

---

## Commercial Marketing: Bulk-Delete Candidates (889 total)

| Sender | Count | Action |
|--------|------:|--------|
| Firebase phishing spam | 103 | **DELETE** |
| Crowdfunding spam farms (6 fake Kickstarter domains) | 99 | **DELETE** |
| Twitch stream notifications | 84 | **DELETE** |
| Zillow instant updates | 67 | **DELETE** |
| The Tree Center marketing | 62 | **DELETE** |
| Freelancer.com notifications | 53 | **DELETE** |
| Meetup notifications | 52 | **DELETE** |
| Facebook noise | 51 | **DELETE** |
| Retail marketing (Domestika/MUJI/TASCHEN/MACK/Baronfig) | 54 | **DELETE** |
| Manning Publications marketing | 49 | **DELETE** |
| Experian credit monitoring | 44 | **DELETE** |
| PledgeBox | 27 | **DELETE** |
| LinkedIn | 28 | **DELETE** |
| ship30for30 | 25 | **DELETE** |
| eBay (non-transactional) | 20 | **DELETE** |
| iPhone Photography School | 22 | **DELETE** |
| Walgreens | 14 | **DELETE** |
| CVS receipts/surveys | 14 | **DELETE** |
| Amazon marketing | 8 | **DELETE** |
| backerclub.co | 8 | **DELETE** |

**Total: 889 messages = 31% of inbox. Safe to bulk-delete.**

---

## Financial & Billing (needs review, not delete)

Real transactional emails to keep / act on:

| Sender | Subject / Action |
|--------|-----------------|
| **Hetzner** (billing@hetzner.com) | ⚠️ Final Payment Warning K1264826925 (March 25) + 2 prior reminders |
| AWS | April invoice available (Account 745667007186) |
| DigitalOcean | March invoice + receipt |
| Apple Card | Statement ready (April 1) + payment due notices |
| Bank of America | Zelle: $2,000 → Emily Mason (March 30), $350 → CLUB CLUB LLC (March 28), $100 → Emily Mason (March 26) |
| PayPal | 3× receipts |
| Stripe | X (Twitter) receipts, Every Media receipt, Chroma receipt |
| LendingClub | Statement + promo spam (discard promo) |
| Rayobyte | 2× invoices (proxy service) |
| Blender Studio | Subscription payment received |
| GitHub | 2× payment receipts (wesen org) |
| Invoice Ninja | Invoice sent to Legato Hearing Inc |
| Cox | Internet bill due (March 20) |
| RI Energy | Electric bill due (March 17 + April 2) |
| Conservice | Utility statement for Halstead Providence |
| Ally Bank | Statement ready |
| Venmo | Feb transaction history; paid Emily Mason $100, $50 |
| Ecamm Network | Payment received |
| Every Media / stripe | Receipt #2564-8750 |

**Note:** There are ~15 fake invoice/payment phishing emails mixed in (random sender addresses, "Your BNH Billing Has Been Finalized", etc.). These use Firebase/random domains and should be discarded.

---

## Security Alerts (REVIEW IMMEDIATELY)

### 🔴 Google — Critical security alert (March 30)
Three emails from `no-reply@accounts.google.com`:
- "Critical security alert for wesen3000@googlemail.com" (02:55 UTC)
- "Security alert for wesen3000@googlemail.com" (02:55 UTC)
- "Security alert for wesen3000@googlemail.com" (10:24 UTC)

**Action:** Log into wesen3000@googlemail.com and review recent activity. Enable 2FA if not set.

### 🔴 Microsoft — Unusual login activity (March 13)
Multiple emails from `account-security-noreply@accountprotection.microsoft.com` (German):
- "Ungewöhnliche Anmeldeaktivität" × 4 (15:18, 15:21, 15:33)
- "Sicherheitsinformation ... hinzugefügt" × 2
- "Ihr Einmalcode"

This pattern (OTP sent, security info added, multiple logins) suggests possible account takeover attempt. **Action:** Review Microsoft account immediately.

### 🟡 Microsoft phishing in inbox
Two emails from `flow-noreply@microsoft.com` are **scam/phishing**:
- "Unusual Transaction USD 859.99 to Pay-Pal BTC Store" 
- "Please review order confirmation 57Q24N"

These are not real Microsoft emails — discard.

### Google security alerts for other accounts
- wesen3000@googlemail.com (multiple, March 18, 20, 30)
- manuel@goldeneaglecoin.com (March 20)

---

## Personal / Human Emails (the 8 that matter)

| Date | From | Subject |
|------|------|---------|
| 2026-03-30 | Gian und Denise Besset Stadler | Re: Ascension à La Forclaz |
| 2026-03-27 | catherinechambel@wanadoo.fr | RE: Ascension à La Forclaz |
| 2026-03-25 | Hans Hübner | Termin abgesagt: Prompter Stammtisch |
| 2026-03-25 | Hans Hübner | Aktualisierte Einladung: Catch up (×2) |
| 2026-03-24 | Hans Hübner | Einladung: Catch up |
| 2026-03-23 | lisa besset | Spotted in the city: Leo a gogo 🐆 |
| 2026-03-07 | Marco Antoniotti | Re: eval-when mysteries |
| 2026-03-03 | Stepan Gantralyan | Vernissage March 6 + Konzert March 14 |

Additional near-personal:
- Manfred Odendahl (family, 4 emails) — account hack + photography
- Invoice Ninja sent to Legato Hearing (your own business)
- 2 community list replies from gmail users

---

## Sieve Rules Summary

File: `scripts/08_sieve_rules.sieve`

**Discard rules cover:** crowdfunding spam, Firebase phishing, Facebook, Instagram, SoundCloud, Zillow, Tree Center, Experian, Freelancer, LinkedIn, Walgreens, CVS, ship30for30, PledgeBox, Twitch (non-billing), retail marketing, Manning (non-transactional), eBay (non-transactional).

**File-to-folder rules cover:** GitHub CI, GitHub PRs, GitHub general, Substack, Beehiiv, Every.to, Readwise, Meetup, mailing lists, Rollbar, W&B, Firecrawl, Replit, Augment Code, CodeRabbit, Apple finance, Bank of America, PayPal, AWS, Stripe, DigitalOcean, Hetzner, LendingClub, Affirm, Google security, Microsoft security, Apple Developer.

**Estimated inbox reduction:** 889 deleted (31%) + ~1,600 filed (55%) = inbox down to ~400 messages/month instead of 2,894 — an **86% reduction**.

---

## Recommended Folder Structure

```
INBOX/                    ← only truly unclassified + personal
github/
  ci/                     ← CI failures (652/month)
  prs/                    ← PR discussions
  notifications/          ← other GitHub
newsletters/
  substack/
  beehiiv/
  every/
  readwise/
  apple-developer/
finance/
  apple/
  bofa/
  paypal/
  aws/
  stripe-receipts/
  digitalocean/
  hetzner/
  lendingclub/
  affirm/
security/
  google/
  microsoft/
monitoring/
  rollbar/
dev-tools/
  wandb/
  firecrawl/
  replit/
  augment/
  coderabbit/
events/
  meetup/
lists/
```

---

## Open Questions

1. Should Rollbar daily summaries be discarded (low signal), and only new errors kept?
2. Should GitHub CI failures for `wesen/goldeneaglecoin.com` trigger a separate alert (prod deployment failures)?
3. Unsubscribe vs. discard: for commercial senders — worth sending unsubscribe requests, or just silently discard?
4. The `go-go-golems/*` "other" repos (382 msgs) — many are likely sub-repos. Worth investigating which repos drive the most noise.
