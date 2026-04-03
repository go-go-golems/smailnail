---
title: Mail Triage Plan
doc_type: design
status: active
intent: long-term
topics:
  - email
  - annotations
  - triage
  - sqlite
  - workflow
owners:
  - manuel
ticket: SMN-20260403-MAIL-TRIAGE
---

# Mail Triage Plan

## Mailbox Profile

- **Total messages:** 32,912 (all INBOX)
- **Date range:** May 2024 – March 2026
- **Unique sender domains:** 1,305
- **Unique sender emails:** 2,187+
- **Threads:** 29,305 (28,405 are single-message)
- **Enrichment:** Complete (threads, senders, unsubscribe links)
- **Existing annotations:** None (clean slate)

## The Problem

~60% of the inbox is automated/notification noise (GitHub CI failures, marketing, commerce). The remaining 40% contains a mix of valuable newsletters, personal correspondence, work discussions, and hobby/creative content — but it's buried. We need to categorize senders and surface the signal.

## Proposed Annotation Taxonomy

### Sender-level tags (applied to `target_type=sender`)

These are the primary triage dimension — classifying each sender once propagates to all their messages.

| Tag | Description | Example senders |
|---|---|---|
| `noise/ci` | CI failure notifications, automated build alerts | notifications@github.com (for CI subset) |
| `noise/marketing` | Marketing, sales, promotions | sales@thetreecenter.com, mkt@manning.com |
| `noise/transactional` | Order confirmations, shipping, receipts | shipment-tracking@amazon.com, auto-confirm@amazon.com |
| `noise/social-notif` | Social platform notifications (not direct messages) | no-reply@twitch.tv, facebookmail.com senders |
| `noise/spam` | Outright spam or low-value bulk | purchasingreviews.com, costsoldier.com, voip-prices.com |
| `newsletter/tech` | Tech newsletters worth reading | substack tech writers, every.to, alphasignal.ai |
| `newsletter/culture` | Non-tech newsletters (art, books, film) | densediscovery.com, charcoalbookclub.com, criterion.com |
| `newsletter/creative` | Music, photography, creative tools | magnumphotos.com, bandcamp.com, plugin-alliance.com |
| `personal` | Real people writing to Manuel directly | gmail.com personal correspondents |
| `work` | Work-related (team-mento, rollbar, slack, zulip) | rollbar, slack, zulip, team-mento GitHub threads |
| `community` | Mailing lists, meetups, hackerspaces | lists.entropia.de, email.meetup.com, recurse (zulip) |
| `financial` | Banking, payments, insurance | paypal.com, bankofamerica.com, affirm.com |
| `services` | Services Manuel uses (housing, car, health) | bozzuto.com, stadtmobil, cvs, walgreens |
| `hobby` | Hobby-related (3d printing, gaming, fitness) | bambulab.com, play.date, itch.io, strava, trainingpeaks |

### Domain-level tags (applied to `target_type=domain`)

For bulk categorization when all senders at a domain share a category.

### Message-level tags (applied to `target_type=message`)

Used sparingly for individual messages that deserve attention:

| Tag | Description |
|---|---|
| `action-required` | Needs a response or action |
| `interesting` | Worth reading, flagged for Manuel's attention |
| `reference` | Keep for future reference |

## Proposed Execution Plan

### Phase 1: Sender categorization (bulk, heuristic)

**Scripts to create:**
1. `01-categorize-noise-senders.sh` — Tag known-noise senders (CI, marketing, spam) based on domain patterns and volume heuristics
2. `02-categorize-newsletter-senders.sh` — Tag newsletter senders (substack, beehiiv, every.to, etc.) with subcategories
3. `03-categorize-personal-senders.sh` — Tag personal correspondents (gmail/icloud/protonmail senders with real conversation patterns)
4. `04-categorize-work-senders.sh` — Tag work-related senders
5. `05-categorize-services-senders.sh` — Tag service/financial/housing senders
6. `06-categorize-hobby-creative.sh` — Tag hobby and creative senders

**Method:** Use `smailnail annotate annotation add` for each sender, with `--source-kind heuristic --source-label mail-triage-v1 --created-by pi-agent`.

**Logging:** Each script will create an `annotation_log` entry summarizing what it did and link it to affected targets.

### Phase 2: Domain-level groups

Create `target_groups` for review clusters:
- "Unsubscribe candidates" — high-volume senders with List-Unsubscribe that Manuel may not read
- "Valuable newsletters" — newsletters that seem worth keeping
- "Personal contacts" — real human correspondents
- "Work threads" — team-mento and related

### Phase 3: Message-level surfacing

After sender categorization, query for interesting uncategorized messages and flag individual standouts:
- Recent personal emails that might need a reply
- Newsletter issues on topics Manuel cares about
- Threads with high engagement (multi-message threads from real people)

### Phase 4: Summary report

Generate a Markdown summary for Manuel:
- "Your inbox at a glance" — category breakdown with counts
- "Action needed" — messages/threads that may need response
- "Worth reading" — interesting newsletters and personal mail
- "Noise to consider unsubscribing from" — high-volume senders with unsubscribe links

## Annotation Conventions

All annotations from this workflow will use:
- `--source-kind heuristic` for rule-based categorization
- `--source-label mail-triage-v1` to identify this run
- `--created-by pi-agent`
- `--review-state to_review` (default for non-human) so Manuel can approve/dismiss

Every script will log its actions via `smailnail annotate log add` with a descriptive title and body, then link targets via `smailnail annotate log link-target`.

## Open Questions for Manuel

1. **Priority override:** Are any of the "noise" senders actually important to you? (e.g., do you read freelancer.com notifications?)
2. **Newsletter preferences:** Which tech newsletters do you actually read vs. let pile up? (The substack list has ~30+ authors)
3. **Work scope:** Is team-mento your current/recent employer? Should all GitHub notifications for team-mento repos be tagged `work`?
4. **Personal contacts:** Any key contacts I should know about beyond what I can infer from gmail/icloud senders?
5. **Hobby priorities:** Which hobby domains matter most? (music production, photography, 3D printing, gaming all seem to be interests)
