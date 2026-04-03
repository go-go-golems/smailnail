## Sieve Rules — generated from inbox analysis 2026-04-02
## Inbox: mail-bl0rg-net-993-manuel (2,894 messages, 2026-03 to 2026-04)
##
## File layout:
##   - DELETE (bulk trash): pure noise with no action value
##   - File to folder: keeps inbox clean, still accessible
## Requires: fileinto, reject/discard, envelope, header extensions

require ["fileinto", "imap4flags", "envelope", "body", "reject"];

## ─────────────────────────────────────────────
## SECTION 1 — DELETE IN BULK (discard)
## These are high-confidence spam / zero-value senders
## ─────────────────────────────────────────────

# Fake crowdfunding spam (kickstarter clone farms)
if anyof (
    address :domain :is "from" "kickstarnow.com",
    address :domain :is "from" "kickstarternew.com",
    address :domain :is "from" "kickstartgenius.com",
    address :domain :is "from" "backerhome.com",
    address :domain :is "from" "kickstarter-new.net",
    address :domain :is "from" "kickstartrend.com",
    address :domain :is "from" "huaweiinsider.com"
) {
    discard;
    stop;
}

# Firebase noreply = phishing / spam
if allof (
    header :contains "from" "firebaseapp.com",
    address :localpart :is "from" "noreply"
) {
    discard;
    stop;
}

# Facebook notifications (friend suggestions, reminders — zero value)
if address :domain :is "from" "facebookmail.com" {
    discard;
    stop;
}

# Instagram unread / stories recap
if address :domain :is "from" "mail.instagram.com" {
    discard;
    stop;
}

# SoundCloud activity alerts
if address :domain :is "from" "notifications.soundcloud.com" {
    discard;
    stop;
}

# Zillow instant updates (came via Apple private relay)
if anyof (
    header :contains "from" "zillow",
    address :domain :is "from" "mail.zillow.com"
) {
    discard;
    stop;
}

# The Tree Center marketing (62 messages/month - pure promo)
if address :domain :is "from" "thetreecenter.com" {
    discard;
    stop;
}

# Experian marketing/credit monitoring spam
if anyof (
    address :domain :is "from" "e.usa.experian.com",
    address :domain :is "from" "s.usa.experian.com"
) {
    discard;
    stop;
}

# Freelancer.com notifications
if address :domain :is "from" "notifications.freelancer.com" {
    discard;
    stop;
}

# LinkedIn (all notifications — nearly no human signal)
if anyof (
    address :domain :is "from" "linkedin.com",
    address :domain :is "from" "notifications-noreply.linkedin.com",
    address :domain :is "from" "messages-noreply.linkedin.com"
) {
    discard;
    stop;
}

# Walgreens
if address :domain :is "from" "eml.walgreens.com" {
    discard;
    stop;
}

# CVS receipts & surveys
if anyof (
    address :domain :is "from" "your.cvs.com",
    address :domain :is "from" "mystore.cvs.com",
    address :domain :is "from" "express.medallia.com"
) {
    discard;
    stop;
}

# ship30for30 / Dickie & Cole writing course spam
if anyof (
    address :domain :is "from" "ship30for30.com"
) {
    discard;
    stop;
}

# Pledge/backer spam
if address :domain :is "from" "pledgebox.com" {
    discard;
    stop;
}

# iPhone Photography School marketing
if address :domain :is "from" "iphonephotographyschool.com" {
    discard;
    stop;
}

# Twitch stream notifications (keep invoice/receipts - handled by billing rule below)
if allof (
    address :domain :is "from" "twitch.tv",
    not subject :contains "invoice",
    not subject :contains "receipt",
    not subject :contains "subscription"
) {
    discard;
    stop;
}

# Retail / lifestyle marketing (MUJI, TASCHEN, MACK, Baronfig, Domestika via relay)
if anyof (
    address :domain :is "from" "muji.us",
    address :domain :is "from" "subscription.taschen-mail.com",
    address :domain :is "from" "mackbooks.co.uk",
    address :domain :is "from" "news.baronfig.com",
    header :contains "from" "domestika"
) {
    discard;
    stop;
}

# KickStarter clone spam (backerclub)
if address :domain :is "from" "backerclub.co" {
    discard;
    stop;
}

# Manning Publications bulk marketing (not purchase confirmations)
if allof (
    address :domain :is "from" "manning.com",
    not subject :contains "order",
    not subject :contains "receipt",
    not subject :contains "invoice"
) {
    discard;
    stop;
}

# eBay marketing (not transaction alerts)
if allof (
    address :domain :is "from" "ebay.com",
    not subject :contains "order",
    not subject :contains "purchase",
    not subject :contains "payment"
) {
    discard;
    stop;
}

## ─────────────────────────────────────────────
## SECTION 2 — FILE TO FOLDER (keep but out of inbox)
## ─────────────────────────────────────────────

# GitHub CI failures → github/ci
if anyof (
    header :contains "subject" "Run failed",
    header :contains "subject" "PR run failed"
) {
    if anyof (
        address :domain :is "from" "github.com",
        header :contains "from" "notifications@github.com"
    ) {
        fileinto "github/ci";
        stop;
    }
}

# GitHub PR discussions / reviews → github/prs
if allof (
    anyof (
        address :domain :is "from" "github.com",
        header :contains "from" "notifications@github.com"
    ),
    header :matches "subject" "Re: [*"
) {
    fileinto "github/prs";
    stop;
}

# All other GitHub notifications → github/notifications
if anyof (
    address :domain :is "from" "github.com",
    header :contains "from" "notifications@github.com",
    header :contains "from" "noreply@github.com"
) {
    fileinto "github/notifications";
    stop;
}

# Substack newsletters → newsletters/substack
if anyof (
    address :domain :is "from" "substack.com",
    address :domain :is "from" "mg1.substack.com"
) {
    fileinto "newsletters/substack";
    stop;
}

# Beehiiv newsletters
if address :domain :is "from" "mail.beehiiv.com" {
    fileinto "newsletters/beehiiv";
    stop;
}

# Every.to newsletter
if address :domain :is "from" "every.to" {
    fileinto "newsletters/every";
    stop;
}

# Readwise
if address :domain :is "from" "readwise.io" {
    fileinto "newsletters/readwise";
    stop;
}

# Meetup event notifications
if address :domain :is "from" "email.meetup.com" {
    fileinto "events/meetup";
    stop;
}

# Mailing lists (riseup, entropia, etc.)
if anyof (
    address :domain :is "from" "lists.riseup.net",
    address :domain :is "from" "lists.entropia.de",
    header :contains "list-id" ""
) {
    fileinto "lists";
    stop;
}

# Rollbar error monitoring → monitoring
if address :domain :is "from" "mail.rollbar.com" {
    fileinto "monitoring/rollbar";
    stop;
}

# W&B / Weights & Biases → dev-tools
if address :domain :is "from" "mail.wandb.ai" {
    fileinto "dev-tools/wandb";
    stop;
}

# Firecrawl → dev-tools
if address :domain :is "from" "firecrawl.dev" {
    fileinto "dev-tools/firecrawl";
    stop;
}

# Replit → dev-tools
if address :domain :is "from" "mail.replit.com" {
    fileinto "dev-tools/replit";
    stop;
}

# Augment Code → dev-tools
if address :domain :is "from" "augmentcode.com" {
    fileinto "dev-tools/augment";
    stop;
}

# CodeRabbit → dev-tools
if address :domain :is "from" "coderabbit.ai" {
    fileinto "dev-tools/coderabbit";
    stop;
}

# Apple Card / receipts → finance/apple
if anyof (
    header :contains "from" "post.applecard.apple",
    allof (address :domain :is "from" "email.apple.com", subject :contains "receipt")
) {
    fileinto "finance/apple";
    stop;
}

# Bank of America → finance/bofa
if address :domain :is "from" "ealerts.bankofamerica.com" {
    fileinto "finance/bofa";
    stop;
}

# PayPal → finance/paypal
if anyof (
    address :domain :is "from" "paypal.com",
    header :contains "from" "service@paypal.com"
) {
    fileinto "finance/paypal";
    stop;
}

# AWS invoices → finance/aws
if anyof (
    address :domain :is "from" "amazonaws.com",
    header :contains "from" "invoicing@aws.com"
) {
    fileinto "finance/aws";
    stop;
}

# Stripe receipts (X, Every, Chroma, etc.)
if header :contains "from" "stripe.com" {
    fileinto "finance/stripe-receipts";
    stop;
}

# DigitalOcean invoices
if address :domain :is "from" "digitalocean.com" {
    fileinto "finance/digitalocean";
    stop;
}

# Hetzner billing
if address :domain :is "from" "hetzner.com" {
    fileinto "finance/hetzner";
    stop;
}

# LendingClub (statements only, promos already discarded above)
if address :domain :is "from" "mail6.lendingclub.com" {
    fileinto "finance/lendingclub";
    stop;
}

# Affirm
if address :domain :is "from" "e.affirm.com" {
    fileinto "finance/affirm";
    stop;
}

# Google security alerts → security
if allof (
    address :domain :is "from" "accounts.google.com",
    anyof (subject :contains "security", subject :contains "alert")
) {
    fileinto "security/google";
    stop;
}

# Microsoft account security
if address :domain :is "from" "accountprotection.microsoft.com" {
    fileinto "security/microsoft";
    stop;
}

# Apple Developer newsletters → newsletters/apple-developer
if address :domain :is "from" "insideapple.apple.com" {
    fileinto "newsletters/apple-developer";
    stop;
}
