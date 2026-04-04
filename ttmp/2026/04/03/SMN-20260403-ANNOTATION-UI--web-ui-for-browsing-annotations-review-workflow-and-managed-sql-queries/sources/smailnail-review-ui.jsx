import { useState, useMemo, useCallback } from "react";

// ─── Mock Data ───────────────────────────────────────────────────────────────

const AGENT_RUNS = [
  { id: "run-47", label: "triage-pass-1", created_at: "2026-04-03T09:12:00Z", annotation_count: 87, reviewed: 0, dismissed: 0 },
  { id: "run-46", label: "newsletter-scanner", created_at: "2026-04-02T14:30:00Z", annotation_count: 34, reviewed: 28, dismissed: 6 },
  { id: "run-45", label: "triage-pass-1", created_at: "2026-03-28T11:00:00Z", annotation_count: 112, reviewed: 108, dismissed: 4 },
  { id: "run-44", label: "bulk-sender-detect", created_at: "2026-03-25T08:45:00Z", annotation_count: 23, reviewed: 23, dismissed: 0 },
  { id: "run-43", label: "triage-pass-1", created_at: "2026-03-20T10:15:00Z", annotation_count: 96, reviewed: 96, dismissed: 0 },
];

const ANNOTATIONS = [
  { id: "a1", target_type: "sender", target_id: "newsletters@medium.com", tag: "newsletter", note: "Sends digest emails 2-3x/week. Has List-Unsubscribe header.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:05Z" },
  { id: "a2", target_type: "sender", target_id: "noreply@github.com", tag: "notification", note: "Automated CI/CD and PR notifications. High volume but relevant.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:08Z" },
  { id: "a3", target_type: "sender", target_id: "deals@shop.example.com", tag: "bulk-sender", note: "Marketing emails. 47 messages, all promotional. Unsubscribe link available.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:12Z" },
  { id: "a4", target_type: "sender", target_id: "hello@buttondown.email", tag: "newsletter", note: "Indie newsletter platform. Delivers subscribed content.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:15Z" },
  { id: "a5", target_type: "thread", target_id: "thread-8812", tag: "important", note: "Contract discussion thread with 4 participants spanning 2 weeks.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:18Z" },
  { id: "a6", target_type: "sender", target_id: "no-reply@accounts.google.com", tag: "transactional", note: "Security alerts and account notifications. Should not be unsubscribed.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:20Z" },
  { id: "a7", target_type: "domain", target_id: "marketing.salesforce.com", tag: "ignore", note: "All messages from this domain are marketing automation. 12 distinct senders.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:24Z" },
  { id: "a8", target_type: "sender", target_id: "weekly@changelog.com", tag: "newsletter", note: "Weekly dev newsletter. Consistent schedule, always has unsubscribe.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:28Z" },
  { id: "a9", target_type: "message", target_id: "msg-44210", tag: "action-required", note: "Contains deadline mention: 'please reply by April 10th'.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:30Z" },
  { id: "a10", target_type: "sender", target_id: "team@linear.app", tag: "notification", note: "Project management notifications. 89 messages in last 30 days.", source_label: "triage-pass-1", agent_run_id: "run-47", review_state: "to_review", created_at: "2026-04-03T09:12:33Z" },
  // Some reviewed/dismissed from older runs
  { id: "a11", target_type: "sender", target_id: "info@substack.com", tag: "newsletter", note: "Newsletter platform delivery.", source_label: "newsletter-scanner", agent_run_id: "run-46", review_state: "reviewed", created_at: "2026-04-02T14:30:10Z" },
  { id: "a12", target_type: "sender", target_id: "noreply@vercel.com", tag: "notification", note: "Build and deployment alerts.", source_label: "newsletter-scanner", agent_run_id: "run-46", review_state: "reviewed", created_at: "2026-04-02T14:30:15Z" },
  { id: "a13", target_type: "sender", target_id: "support@fly.io", tag: "newsletter", note: "Misclassified — these are support ticket replies.", source_label: "newsletter-scanner", agent_run_id: "run-46", review_state: "dismissed", created_at: "2026-04-02T14:30:20Z" },
];

const GROUPS = [
  { id: "g1", name: "Possible Newsletters", description: "Senders that exhibit newsletter patterns: regular cadence, unsubscribe headers, one-to-many delivery.", agent_run_id: "run-47", review_state: "to_review", member_count: 6,
    members: [
      { target_type: "sender", target_id: "newsletters@medium.com" },
      { target_type: "sender", target_id: "hello@buttondown.email" },
      { target_type: "sender", target_id: "weekly@changelog.com" },
      { target_type: "sender", target_id: "info@substack.com" },
      { target_type: "sender", target_id: "digest@hackernews.com" },
      { target_type: "sender", target_id: "weekly@tldr.tech" },
    ]
  },
  { id: "g2", name: "High-Volume Notification Senders", description: "Automated senders producing 50+ messages/month. Mostly CI, project management, and monitoring.", agent_run_id: "run-47", review_state: "to_review", member_count: 4,
    members: [
      { target_type: "sender", target_id: "noreply@github.com" },
      { target_type: "sender", target_id: "team@linear.app" },
      { target_type: "sender", target_id: "alerts@datadog.com" },
      { target_type: "sender", target_id: "builds@circleci.com" },
    ]
  },
  { id: "g3", name: "Marketing / Ignorable", description: "Pure marketing senders. Safe to bulk-ignore or unsubscribe.", agent_run_id: "run-47", review_state: "to_review", member_count: 3,
    members: [
      { target_type: "sender", target_id: "deals@shop.example.com" },
      { target_type: "sender", target_id: "promo@store.example.com" },
      { target_type: "sender", target_id: "offers@brand.example.com" },
    ]
  },
];

const LOGS = [
  { id: "l1", log_kind: "run-start", title: "Triage pass started", body: "Beginning triage of 342 unseen messages across 3 accounts. Focus: sender classification and thread importance.", agent_run_id: "run-47", created_at: "2026-04-03T09:12:00Z",
    targets: [] },
  { id: "l2", log_kind: "note", title: "Newsletter pattern detected", body: "Found 6 senders matching newsletter heuristics:\n- Regular sending cadence (weekly or more)\n- Presence of List-Unsubscribe header\n- One-to-many delivery pattern (no personalized replies)\n\nGrouped as 'Possible Newsletters' for batch review.", agent_run_id: "run-47", created_at: "2026-04-03T09:12:15Z",
    targets: [{ type: "sender", id: "newsletters@medium.com" }, { type: "sender", id: "hello@buttondown.email" }] },
  { id: "l3", log_kind: "note", title: "High-volume senders flagged", body: "4 senders exceed 50 msgs/month. All appear to be automated notifications (CI, PM tools, monitoring). Tagged as `notification` rather than `bulk-sender` because content is contextually relevant to the operator's work.", agent_run_id: "run-47", created_at: "2026-04-03T09:12:25Z",
    targets: [{ type: "sender", id: "noreply@github.com" }, { type: "sender", id: "team@linear.app" }] },
  { id: "l4", log_kind: "note", title: "Domain-level ignore recommendation", body: "marketing.salesforce.com has 12 distinct sender addresses, all sending promotional content. Recommending domain-level `ignore` tag rather than individual sender tags to reduce annotation noise.", agent_run_id: "run-47", created_at: "2026-04-03T09:12:30Z",
    targets: [{ type: "domain", id: "marketing.salesforce.com" }] },
  { id: "l5", log_kind: "run-end", title: "Triage pass complete", body: "Processed 342 messages. Created 87 annotations across 10 senders, 1 thread, 1 domain, 1 message. 3 target groups created for batch review. Estimated review time: 5-8 minutes.", agent_run_id: "run-47", created_at: "2026-04-03T09:12:40Z",
    targets: [] },
];

const SENDERS = [
  { email: "newsletters@medium.com", display_name: "Medium Daily Digest", domain: "medium.com", msg_count: 156, first_seen: "2024-01-15", last_seen: "2026-04-02", has_unsubscribe: true },
  { email: "noreply@github.com", display_name: "GitHub", domain: "github.com", msg_count: 2341, first_seen: "2023-06-01", last_seen: "2026-04-03", has_unsubscribe: true },
  { email: "deals@shop.example.com", display_name: "Shop Deals", domain: "shop.example.com", msg_count: 47, first_seen: "2025-08-20", last_seen: "2026-04-01", has_unsubscribe: true },
  { email: "hello@buttondown.email", display_name: "Buttondown", domain: "buttondown.email", msg_count: 38, first_seen: "2025-02-10", last_seen: "2026-03-30", has_unsubscribe: true },
  { email: "no-reply@accounts.google.com", display_name: "Google Accounts", domain: "accounts.google.com", msg_count: 24, first_seen: "2023-06-01", last_seen: "2026-03-28", has_unsubscribe: false },
  { email: "weekly@changelog.com", display_name: "Changelog Weekly", domain: "changelog.com", msg_count: 102, first_seen: "2024-03-01", last_seen: "2026-04-01", has_unsubscribe: true },
  { email: "team@linear.app", display_name: "Linear", domain: "linear.app", msg_count: 89, first_seen: "2025-01-10", last_seen: "2026-04-03", has_unsubscribe: true },
];

const PRESET_QUERIES = [
  { name: "top-senders.sql", description: "Top senders by message count with annotation coverage", sql: `SELECT s.email, s.domain, s.msg_count,\n  COUNT(a.id) AS annotations\nFROM senders s\nLEFT JOIN annotations a\n  ON a.target_type = 'sender'\n  AND a.target_id = s.email\nGROUP BY s.email\nORDER BY s.msg_count DESC\nLIMIT 50;` },
  { name: "unreviewed-by-tag.sql", description: "Count of unreviewed annotations grouped by tag", sql: `SELECT tag, COUNT(*) as cnt\nFROM annotations\nWHERE review_state = 'to_review'\nGROUP BY tag\nORDER BY cnt DESC;` },
  { name: "recent-threads.sql", description: "Threads with recent activity and annotation state", sql: `SELECT t.thread_id, t.subject,\n  t.message_count, t.last_sent_date,\n  COUNT(a.id) AS annotations\nFROM threads t\nLEFT JOIN annotations a\n  ON a.target_type = 'thread'\n  AND a.target_id = t.thread_id\nWHERE t.last_sent_date > date('now', '-7 days')\nGROUP BY t.thread_id\nORDER BY t.last_sent_date DESC;` },
  { name: "domain-summary.sql", description: "Domains ranked by total message volume", sql: `SELECT sender_domain, COUNT(*) as msgs,\n  COUNT(DISTINCT sender_email) as senders\nFROM messages\nGROUP BY sender_domain\nORDER BY msgs DESC\nLIMIT 30;` },
];

// ─── Utility ─────────────────────────────────────────────────────────────────

const fmt = {
  date: (s) => {
    if (!s) return "—";
    const d = new Date(s);
    return d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
  },
  time: (s) => {
    if (!s) return "";
    const d = new Date(s);
    return d.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" });
  },
  relative: (s) => {
    const d = new Date(s);
    const now = new Date("2026-04-03T15:00:00Z");
    const diff = now - d;
    const mins = Math.floor(diff / 60000);
    if (mins < 60) return `${mins}m ago`;
    const hrs = Math.floor(mins / 60);
    if (hrs < 24) return `${hrs}h ago`;
    const days = Math.floor(hrs / 24);
    return `${days}d ago`;
  },
  num: (n) => n?.toLocaleString() ?? "—",
};

// ─── Styles ──────────────────────────────────────────────────────────────────

const palette = {
  bg: "#0c0e11",
  bgRaised: "#13161b",
  bgHover: "#1a1e25",
  bgActive: "#1f2530",
  border: "#252a33",
  borderLight: "#2e3440",
  text: "#c8cdd5",
  textMuted: "#6b7280",
  textDim: "#4a5060",
  textBright: "#e8ecf2",
  accent: "#d4a053",
  accentDim: "#a07830",
  green: "#4ade80",
  greenDim: "#166534",
  red: "#f87171",
  redDim: "#7f1d1d",
  blue: "#60a5fa",
  purple: "#a78bfa",
  orange: "#fb923c",
  tag: {
    newsletter: { bg: "#1e293b", fg: "#60a5fa", border: "#334155" },
    notification: { bg: "#1c1917", fg: "#fb923c", border: "#3f3224" },
    "bulk-sender": { bg: "#1a1625", fg: "#a78bfa", border: "#2e2544" },
    important: { bg: "#14231a", fg: "#4ade80", border: "#1e3a28" },
    transactional: { bg: "#1a1a14", fg: "#d4a053", border: "#33301e" },
    ignore: { bg: "#1c1115", fg: "#f87171", border: "#3b1a22" },
    "action-required": { bg: "#231414", fg: "#f87171", border: "#3b1a1a" },
  },
};

const font = {
  mono: "'JetBrains Mono', 'Fira Code', 'SF Mono', 'Cascadia Code', monospace",
  body: "'IBM Plex Sans', -apple-system, sans-serif",
};

const S = {
  app: { display: "flex", height: "100vh", width: "100vw", background: palette.bg, color: palette.text, fontFamily: font.body, fontSize: 13, overflow: "hidden" },
  sidebar: { width: 220, minWidth: 220, borderRight: `1px solid ${palette.border}`, display: "flex", flexDirection: "column", background: palette.bg },
  sidebarHead: { padding: "16px 16px 12px", borderBottom: `1px solid ${palette.border}` },
  sidebarLogo: { fontFamily: font.mono, fontSize: 14, fontWeight: 700, color: palette.accent, letterSpacing: "0.05em", margin: 0 },
  sidebarSub: { fontSize: 10, color: palette.textDim, marginTop: 2, fontFamily: font.mono, letterSpacing: "0.08em", textTransform: "uppercase" },
  navSection: { padding: "12px 0 4px 16px", fontSize: 10, fontWeight: 600, color: palette.textDim, textTransform: "uppercase", letterSpacing: "0.1em", fontFamily: font.mono },
  navItem: (active) => ({
    display: "flex", alignItems: "center", gap: 8, padding: "6px 16px", cursor: "pointer", fontSize: 13,
    background: active ? palette.bgActive : "transparent", color: active ? palette.textBright : palette.textMuted,
    borderRight: active ? `2px solid ${palette.accent}` : "2px solid transparent",
    transition: "all 0.15s",
  }),
  navBadge: { marginLeft: "auto", background: palette.accentDim, color: palette.accent, fontSize: 10, fontWeight: 700, padding: "1px 6px", borderRadius: 3, fontFamily: font.mono },
  main: { flex: 1, display: "flex", flexDirection: "column", overflow: "hidden" },
  topBar: { display: "flex", alignItems: "center", gap: 12, padding: "10px 20px", borderBottom: `1px solid ${palette.border}`, minHeight: 44 },
  topTitle: { fontSize: 14, fontWeight: 600, color: palette.textBright, fontFamily: font.mono },
  topMeta: { fontSize: 11, color: palette.textDim },
  content: { flex: 1, overflow: "auto", padding: 20 },
  card: { background: palette.bgRaised, border: `1px solid ${palette.border}`, borderRadius: 4, marginBottom: 12 },
  cardHead: { display: "flex", alignItems: "center", justifyContent: "space-between", padding: "10px 14px", borderBottom: `1px solid ${palette.border}` },
  cardTitle: { fontSize: 12, fontWeight: 600, color: palette.textBright, fontFamily: font.mono },
  cardBody: { padding: 14 },
  table: { width: "100%", borderCollapse: "collapse", fontSize: 12 },
  th: { textAlign: "left", padding: "6px 10px", borderBottom: `1px solid ${palette.border}`, color: palette.textDim, fontWeight: 600, fontSize: 10, textTransform: "uppercase", letterSpacing: "0.08em", fontFamily: font.mono },
  td: { padding: "8px 10px", borderBottom: `1px solid ${palette.border}22`, verticalAlign: "top" },
  tag: (type) => {
    const t = palette.tag[type] || { bg: palette.bgActive, fg: palette.textMuted, border: palette.border };
    return { display: "inline-block", padding: "2px 8px", borderRadius: 3, fontSize: 11, fontWeight: 600, fontFamily: font.mono, background: t.bg, color: t.fg, border: `1px solid ${t.border}` };
  },
  reviewBadge: (state) => ({
    display: "inline-block", padding: "2px 8px", borderRadius: 3, fontSize: 10, fontWeight: 700, fontFamily: font.mono, textTransform: "uppercase", letterSpacing: "0.05em",
    background: state === "to_review" ? palette.accentDim : state === "reviewed" ? palette.greenDim : palette.redDim,
    color: state === "to_review" ? palette.accent : state === "reviewed" ? palette.green : palette.red,
  }),
  btn: (variant = "default") => ({
    display: "inline-flex", alignItems: "center", gap: 4, padding: "4px 10px", borderRadius: 3, fontSize: 11, fontWeight: 600, cursor: "pointer", border: "1px solid", transition: "all 0.15s", fontFamily: font.mono,
    ...(variant === "approve" ? { background: palette.greenDim, color: palette.green, borderColor: "#1e4d2e" } :
       variant === "dismiss" ? { background: palette.redDim, color: palette.red, borderColor: "#4d1e1e" } :
       variant === "primary" ? { background: palette.accentDim, color: palette.accent, borderColor: "#6b5020" } :
       { background: palette.bgActive, color: palette.textMuted, borderColor: palette.border }),
  }),
  code: { fontFamily: font.mono, fontSize: 12, background: palette.bgActive, padding: "1px 5px", borderRadius: 2, color: palette.accent },
  mono: { fontFamily: font.mono, fontSize: 12 },
  link: { color: palette.blue, cursor: "pointer", textDecoration: "none", fontSize: 12 },
  pill: { display: "inline-flex", alignItems: "center", gap: 4, padding: "3px 8px", borderRadius: 12, fontSize: 11, background: palette.bgActive, color: palette.textMuted, border: `1px solid ${palette.border}` },
  emptyState: { textAlign: "center", padding: 40, color: palette.textDim, fontFamily: font.mono, fontSize: 12 },
};

// ─── Components ──────────────────────────────────────────────────────────────

function Tag({ type }) {
  return <span style={S.tag(type)}>{type}</span>;
}

function ReviewBadge({ state }) {
  const labels = { to_review: "To Review", reviewed: "Reviewed", dismissed: "Dismissed" };
  return <span style={S.reviewBadge(state)}>{labels[state]}</span>;
}

function Btn({ children, variant, onClick, style: sx }) {
  return <button style={{ ...S.btn(variant), ...sx }} onClick={onClick}>{children}</button>;
}

function TargetLink({ type, id, onClick }) {
  const icons = { sender: "◉", thread: "⇶", message: "✉", domain: "◈", mailbox: "▤", account: "⊞" };
  return (
    <span style={S.link} onClick={onClick}>
      <span style={{ marginRight: 4, opacity: 0.6 }}>{icons[type] || "·"}</span>
      <span style={S.mono}>{id}</span>
    </span>
  );
}

function StatBox({ label, value, color }) {
  return (
    <div style={{ textAlign: "center", padding: "8px 16px" }}>
      <div style={{ fontSize: 22, fontWeight: 700, fontFamily: font.mono, color: color || palette.textBright }}>{value}</div>
      <div style={{ fontSize: 10, color: palette.textDim, textTransform: "uppercase", letterSpacing: "0.08em", fontFamily: font.mono, marginTop: 2 }}>{label}</div>
    </div>
  );
}

// ─── Views ───────────────────────────────────────────────────────────────────

function ReviewQueueView({ annotations, setAnnotations, navigate }) {
  const pending = annotations.filter(a => a.review_state === "to_review");
  const [selected, setSelected] = useState(new Set());
  const [filterTag, setFilterTag] = useState(null);
  const [filterType, setFilterType] = useState(null);

  const filtered = pending.filter(a =>
    (!filterTag || a.tag === filterTag) && (!filterType || a.target_type === filterType)
  );
  const tags = [...new Set(pending.map(a => a.tag))];
  const types = [...new Set(pending.map(a => a.target_type))];

  const toggleSelect = (id) => setSelected(prev => {
    const next = new Set(prev);
    next.has(id) ? next.delete(id) : next.add(id);
    return next;
  });
  const selectAll = () => setSelected(new Set(filtered.map(a => a.id)));
  const selectNone = () => setSelected(new Set());

  const batchAction = (state) => {
    setAnnotations(prev => prev.map(a => selected.has(a.id) ? { ...a, review_state: state } : a));
    setSelected(new Set());
  };

  return (
    <div>
      <div style={{ display: "flex", gap: 16, marginBottom: 16, flexWrap: "wrap", alignItems: "center" }}>
        <div style={{ display: "flex", gap: 6, alignItems: "center" }}>
          <span style={{ fontSize: 11, color: palette.textDim, fontFamily: font.mono }}>TAG:</span>
          <span style={{ ...S.pill, cursor: "pointer", background: !filterTag ? palette.bgHover : palette.bgActive }} onClick={() => setFilterTag(null)}>all</span>
          {tags.map(t => (
            <span key={t} style={{ ...S.pill, cursor: "pointer", background: filterTag === t ? palette.bgHover : palette.bgActive }} onClick={() => setFilterTag(filterTag === t ? null : t)}>
              {t}
            </span>
          ))}
        </div>
        <div style={{ display: "flex", gap: 6, alignItems: "center" }}>
          <span style={{ fontSize: 11, color: palette.textDim, fontFamily: font.mono }}>TYPE:</span>
          <span style={{ ...S.pill, cursor: "pointer", background: !filterType ? palette.bgHover : palette.bgActive }} onClick={() => setFilterType(null)}>all</span>
          {types.map(t => (
            <span key={t} style={{ ...S.pill, cursor: "pointer", background: filterType === t ? palette.bgHover : palette.bgActive }} onClick={() => setFilterType(filterType === t ? null : t)}>
              {t}
            </span>
          ))}
        </div>
      </div>

      {filtered.length === 0 ? (
        <div style={S.emptyState}>✓ No pending annotations. Queue is clear.</div>
      ) : (
        <div style={S.card}>
          <div style={{ ...S.cardHead, gap: 8 }}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <input type="checkbox" checked={selected.size === filtered.length && filtered.length > 0} onChange={() => selected.size === filtered.length ? selectNone() : selectAll()} />
              <span style={{ fontSize: 11, color: palette.textDim, fontFamily: font.mono }}>
                {selected.size > 0 ? `${selected.size} selected` : `${filtered.length} pending`}
              </span>
            </div>
            {selected.size > 0 && (
              <div style={{ display: "flex", gap: 6 }}>
                <Btn variant="approve" onClick={() => batchAction("reviewed")}>✓ Approve ({selected.size})</Btn>
                <Btn variant="dismiss" onClick={() => batchAction("dismissed")}>✗ Dismiss ({selected.size})</Btn>
              </div>
            )}
          </div>
          <table style={S.table}>
            <thead>
              <tr>
                <th style={{ ...S.th, width: 30 }}></th>
                <th style={S.th}>Target</th>
                <th style={S.th}>Tag</th>
                <th style={S.th}>Agent Note</th>
                <th style={S.th}>Source</th>
                <th style={{ ...S.th, width: 120 }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(a => (
                <tr key={a.id} style={{ background: selected.has(a.id) ? palette.bgActive + "80" : "transparent" }}>
                  <td style={S.td}>
                    <input type="checkbox" checked={selected.has(a.id)} onChange={() => toggleSelect(a.id)} />
                  </td>
                  <td style={S.td}>
                    <TargetLink type={a.target_type} id={a.target_id} onClick={() => {
                      if (a.target_type === "sender") navigate("sender-detail", { email: a.target_id });
                    }} />
                  </td>
                  <td style={S.td}><Tag type={a.tag} /></td>
                  <td style={{ ...S.td, maxWidth: 320, color: palette.textMuted, fontSize: 12, lineHeight: 1.4 }}>{a.note}</td>
                  <td style={S.td}><span style={S.code}>{a.source_label}</span></td>
                  <td style={S.td}>
                    <div style={{ display: "flex", gap: 4 }}>
                      <Btn variant="approve" onClick={() => setAnnotations(prev => prev.map(x => x.id === a.id ? { ...x, review_state: "reviewed" } : x))}>✓</Btn>
                      <Btn variant="dismiss" onClick={() => setAnnotations(prev => prev.map(x => x.id === a.id ? { ...x, review_state: "dismissed" } : x))}>✗</Btn>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function AgentRunsView({ navigate }) {
  return (
    <div>
      <div style={S.card}>
        <table style={S.table}>
          <thead>
            <tr>
              <th style={S.th}>Run</th>
              <th style={S.th}>Agent</th>
              <th style={S.th}>When</th>
              <th style={S.th}>Annotations</th>
              <th style={S.th}>Progress</th>
              <th style={S.th}></th>
            </tr>
          </thead>
          <tbody>
            {AGENT_RUNS.map(r => {
              const done = r.reviewed + r.dismissed;
              const pct = r.annotation_count > 0 ? Math.round((done / r.annotation_count) * 100) : 0;
              return (
                <tr key={r.id}>
                  <td style={S.td}><span style={S.code}>{r.id}</span></td>
                  <td style={S.td}><span style={{ fontFamily: font.mono, fontSize: 12, color: palette.textBright }}>{r.label}</span></td>
                  <td style={S.td}>
                    <div style={{ fontSize: 12 }}>{fmt.date(r.created_at)}</div>
                    <div style={{ fontSize: 10, color: palette.textDim }}>{fmt.relative(r.created_at)}</div>
                  </td>
                  <td style={S.td}>
                    <span style={{ fontFamily: font.mono, fontSize: 14, fontWeight: 700, color: palette.textBright }}>{r.annotation_count}</span>
                    <span style={{ fontSize: 10, color: palette.textDim, marginLeft: 6 }}>
                      {r.reviewed > 0 && <span style={{ color: palette.green }}>✓{r.reviewed}</span>}
                      {r.dismissed > 0 && <span style={{ color: palette.red, marginLeft: 4 }}>✗{r.dismissed}</span>}
                    </span>
                  </td>
                  <td style={S.td}>
                    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                      <div style={{ flex: 1, height: 4, background: palette.bgActive, borderRadius: 2, overflow: "hidden", maxWidth: 100 }}>
                        <div style={{ width: `${pct}%`, height: "100%", background: pct === 100 ? palette.green : palette.accent, borderRadius: 2, transition: "width 0.3s" }} />
                      </div>
                      <span style={{ fontSize: 10, color: palette.textDim, fontFamily: font.mono }}>{pct}%</span>
                    </div>
                  </td>
                  <td style={S.td}>
                    <Btn onClick={() => navigate("run-detail", { runId: r.id })}>Inspect →</Btn>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function RunDetailView({ runId, annotations, setAnnotations, navigate }) {
  const runAnnotations = annotations.filter(a => a.agent_run_id === runId);
  const runLogs = LOGS.filter(l => l.agent_run_id === runId);
  const runGroups = GROUPS.filter(g => g.agent_run_id === runId);
  const run = AGENT_RUNS.find(r => r.id === runId);

  const pending = runAnnotations.filter(a => a.review_state === "to_review");
  const reviewed = runAnnotations.filter(a => a.review_state === "reviewed");
  const dismissed = runAnnotations.filter(a => a.review_state === "dismissed");

  const approveAll = () => setAnnotations(prev => prev.map(a => a.agent_run_id === runId && a.review_state === "to_review" ? { ...a, review_state: "reviewed" } : a));

  return (
    <div>
      {/* Stats row */}
      <div style={{ display: "flex", gap: 1, marginBottom: 16 }}>
        {[
          { label: "Total", value: runAnnotations.length, color: palette.textBright },
          { label: "Pending", value: pending.length, color: palette.accent },
          { label: "Approved", value: reviewed.length, color: palette.green },
          { label: "Dismissed", value: dismissed.length, color: palette.red },
        ].map((s, i) => (
          <div key={i} style={{ flex: 1, background: palette.bgRaised, border: `1px solid ${palette.border}`, padding: "12px 8px", borderRadius: i === 0 ? "4px 0 0 4px" : i === 3 ? "0 4px 4px 0" : 0 }}>
            <StatBox {...s} />
          </div>
        ))}
      </div>

      {/* Batch actions */}
      {pending.length > 0 && (
        <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
          <Btn variant="approve" onClick={approveAll}>✓ Approve All Pending ({pending.length})</Btn>
        </div>
      )}

      {/* Logs timeline */}
      <div style={S.card}>
        <div style={S.cardHead}><span style={S.cardTitle}>Agent Log</span></div>
        <div style={{ padding: "8px 14px" }}>
          {runLogs.map((log, i) => (
            <div key={log.id} style={{ display: "flex", gap: 12, padding: "10px 0", borderBottom: i < runLogs.length - 1 ? `1px solid ${palette.border}22` : "none" }}>
              <div style={{ width: 60, flexShrink: 0, textAlign: "right" }}>
                <div style={{ fontSize: 10, color: palette.textDim, fontFamily: font.mono }}>{fmt.time(log.created_at)}</div>
                <div style={{ fontSize: 9, color: palette.textDim, fontFamily: font.mono, marginTop: 2, textTransform: "uppercase",
                  color: log.log_kind === "run-start" ? palette.green : log.log_kind === "run-end" ? palette.accent : palette.blue,
                }}>{log.log_kind}</div>
              </div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 12, fontWeight: 600, color: palette.textBright, marginBottom: 4 }}>{log.title}</div>
                <div style={{ fontSize: 12, color: palette.textMuted, lineHeight: 1.5, whiteSpace: "pre-wrap" }}>{log.body}</div>
                {log.targets.length > 0 && (
                  <div style={{ display: "flex", gap: 8, marginTop: 6, flexWrap: "wrap" }}>
                    {log.targets.map((t, j) => (
                      <TargetLink key={j} type={t.type} id={t.id} onClick={() => {
                        if (t.type === "sender") navigate("sender-detail", { email: t.id });
                      }} />
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Groups */}
      {runGroups.length > 0 && (
        <div style={S.card}>
          <div style={S.cardHead}><span style={S.cardTitle}>Target Groups</span></div>
          <div style={{ padding: "8px 14px" }}>
            {runGroups.map(g => (
              <div key={g.id} style={{ padding: "10px 0", borderBottom: `1px solid ${palette.border}22` }}>
                <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
                  <span style={{ fontSize: 13, fontWeight: 600, color: palette.textBright }}>{g.name}</span>
                  <ReviewBadge state={g.review_state} />
                  <span style={S.pill}>{g.member_count} members</span>
                </div>
                <div style={{ fontSize: 12, color: palette.textMuted, marginBottom: 8 }}>{g.description}</div>
                <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
                  {g.members.map((m, j) => (
                    <TargetLink key={j} type={m.target_type} id={m.target_id} onClick={() => {
                      if (m.target_type === "sender") navigate("sender-detail", { email: m.target_id });
                    }} />
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Annotations list */}
      <div style={S.card}>
        <div style={S.cardHead}><span style={S.cardTitle}>Annotations ({runAnnotations.length})</span></div>
        <table style={S.table}>
          <thead>
            <tr>
              <th style={S.th}>Target</th>
              <th style={S.th}>Tag</th>
              <th style={S.th}>Note</th>
              <th style={S.th}>State</th>
              <th style={{ ...S.th, width: 100 }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {runAnnotations.map(a => (
              <tr key={a.id}>
                <td style={S.td}>
                  <TargetLink type={a.target_type} id={a.target_id} onClick={() => {
                    if (a.target_type === "sender") navigate("sender-detail", { email: a.target_id });
                  }} />
                </td>
                <td style={S.td}><Tag type={a.tag} /></td>
                <td style={{ ...S.td, maxWidth: 300, color: palette.textMuted, fontSize: 12 }}>{a.note}</td>
                <td style={S.td}><ReviewBadge state={a.review_state} /></td>
                <td style={S.td}>
                  {a.review_state === "to_review" && (
                    <div style={{ display: "flex", gap: 4 }}>
                      <Btn variant="approve" onClick={() => setAnnotations(prev => prev.map(x => x.id === a.id ? { ...x, review_state: "reviewed" } : x))}>✓</Btn>
                      <Btn variant="dismiss" onClick={() => setAnnotations(prev => prev.map(x => x.id === a.id ? { ...x, review_state: "dismissed" } : x))}>✗</Btn>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function SendersView({ annotations, navigate }) {
  return (
    <div>
      <div style={S.card}>
        <table style={S.table}>
          <thead>
            <tr>
              <th style={S.th}>Sender</th>
              <th style={S.th}>Domain</th>
              <th style={S.th}>Messages</th>
              <th style={S.th}>Last Seen</th>
              <th style={S.th}>Tags</th>
              <th style={S.th}>Unsub</th>
            </tr>
          </thead>
          <tbody>
            {SENDERS.map(s => {
              const senderAnnotations = annotations.filter(a => a.target_type === "sender" && a.target_id === s.email);
              return (
                <tr key={s.email} style={{ cursor: "pointer" }} onClick={() => navigate("sender-detail", { email: s.email })}>
                  <td style={S.td}>
                    <div style={{ fontSize: 12, fontWeight: 600, color: palette.textBright }}>{s.display_name}</div>
                    <div style={{ fontSize: 11, color: palette.textDim, fontFamily: font.mono }}>{s.email}</div>
                  </td>
                  <td style={S.td}><span style={S.code}>{s.domain}</span></td>
                  <td style={S.td}><span style={{ fontFamily: font.mono, fontWeight: 600 }}>{fmt.num(s.msg_count)}</span></td>
                  <td style={S.td}><span style={{ fontSize: 11, color: palette.textDim }}>{fmt.date(s.last_seen)}</span></td>
                  <td style={S.td}>
                    <div style={{ display: "flex", gap: 4, flexWrap: "wrap" }}>
                      {senderAnnotations.map(a => <Tag key={a.id} type={a.tag} />)}
                      {senderAnnotations.length === 0 && <span style={{ fontSize: 11, color: palette.textDim }}>—</span>}
                    </div>
                  </td>
                  <td style={S.td}>
                    {s.has_unsubscribe ? <span style={{ color: palette.green, fontSize: 11 }}>●</span> : <span style={{ color: palette.textDim, fontSize: 11 }}>○</span>}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function SenderDetailView({ email, annotations, setAnnotations, navigate }) {
  const sender = SENDERS.find(s => s.email === email);
  const senderAnnotations = annotations.filter(a => a.target_type === "sender" && a.target_id === email);
  const relatedLogs = LOGS.filter(l => l.targets.some(t => t.type === "sender" && t.id === email));

  if (!sender) return <div style={S.emptyState}>Sender not found: {email}</div>;

  return (
    <div>
      {/* Header */}
      <div style={{ ...S.card, padding: 16 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
          <div>
            <h2 style={{ margin: 0, fontSize: 16, fontWeight: 700, color: palette.textBright }}>{sender.display_name}</h2>
            <div style={{ fontFamily: font.mono, fontSize: 13, color: palette.accent, marginTop: 4 }}>{sender.email}</div>
          </div>
          <div style={{ display: "flex", gap: 4 }}>
            {senderAnnotations.map(a => <Tag key={a.id} type={a.tag} />)}
          </div>
        </div>
        <div style={{ display: "flex", gap: 24, marginTop: 16 }}>
          <StatBox label="Messages" value={fmt.num(sender.msg_count)} />
          <StatBox label="Domain" value={sender.domain} color={palette.blue} />
          <StatBox label="First Seen" value={fmt.date(sender.first_seen)} />
          <StatBox label="Last Seen" value={fmt.date(sender.last_seen)} />
          <StatBox label="Unsubscribe" value={sender.has_unsubscribe ? "Available" : "None"} color={sender.has_unsubscribe ? palette.green : palette.textDim} />
        </div>
      </div>

      {/* Annotations */}
      <div style={S.card}>
        <div style={S.cardHead}><span style={S.cardTitle}>Annotations ({senderAnnotations.length})</span></div>
        {senderAnnotations.length === 0 ? (
          <div style={{ ...S.emptyState, padding: 20 }}>No annotations for this sender.</div>
        ) : (
          <table style={S.table}>
            <thead>
              <tr>
                <th style={S.th}>Tag</th>
                <th style={S.th}>Note</th>
                <th style={S.th}>Source</th>
                <th style={S.th}>State</th>
                <th style={S.th}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {senderAnnotations.map(a => (
                <tr key={a.id}>
                  <td style={S.td}><Tag type={a.tag} /></td>
                  <td style={{ ...S.td, color: palette.textMuted, fontSize: 12 }}>{a.note}</td>
                  <td style={S.td}><span style={S.code}>{a.source_label}</span></td>
                  <td style={S.td}><ReviewBadge state={a.review_state} /></td>
                  <td style={S.td}>
                    {a.review_state === "to_review" && (
                      <div style={{ display: "flex", gap: 4 }}>
                        <Btn variant="approve" onClick={() => setAnnotations(prev => prev.map(x => x.id === a.id ? { ...x, review_state: "reviewed" } : x))}>✓</Btn>
                        <Btn variant="dismiss" onClick={() => setAnnotations(prev => prev.map(x => x.id === a.id ? { ...x, review_state: "dismissed" } : x))}>✗</Btn>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Related logs */}
      {relatedLogs.length > 0 && (
        <div style={S.card}>
          <div style={S.cardHead}><span style={S.cardTitle}>Agent Reasoning</span></div>
          <div style={{ padding: "8px 14px" }}>
            {relatedLogs.map(log => (
              <div key={log.id} style={{ padding: "10px 0", borderBottom: `1px solid ${palette.border}22` }}>
                <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
                  <span style={{ fontSize: 12, fontWeight: 600, color: palette.textBright }}>{log.title}</span>
                  <span style={{ fontSize: 10, color: palette.textDim, fontFamily: font.mono }}>{fmt.relative(log.created_at)}</span>
                </div>
                <div style={{ fontSize: 12, color: palette.textMuted, lineHeight: 1.5, whiteSpace: "pre-wrap" }}>{log.body}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Mock recent messages */}
      <div style={S.card}>
        <div style={S.cardHead}><span style={S.cardTitle}>Recent Messages</span></div>
        <table style={S.table}>
          <thead>
            <tr>
              <th style={S.th}>Date</th>
              <th style={S.th}>Subject</th>
              <th style={S.th}>Size</th>
            </tr>
          </thead>
          <tbody>
            {[
              { date: "2026-04-02", subject: "Your Daily Read — AI & Design", size: "14KB" },
              { date: "2026-03-30", subject: "Weekly Digest: Top Stories", size: "18KB" },
              { date: "2026-03-27", subject: "This Week in Technology", size: "12KB" },
              { date: "2026-03-23", subject: "Your Daily Read — Open Source", size: "15KB" },
              { date: "2026-03-20", subject: "Weekly Digest: Dev Tools", size: "16KB" },
            ].map((m, i) => (
              <tr key={i}>
                <td style={S.td}><span style={{ fontSize: 11, color: palette.textDim }}>{fmt.date(m.date)}</span></td>
                <td style={S.td}><span style={{ fontSize: 12, color: palette.textBright }}>{m.subject}</span></td>
                <td style={S.td}><span style={{ fontSize: 11, color: palette.textDim, fontFamily: font.mono }}>{m.size}</span></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function GroupsView({ navigate }) {
  return (
    <div>
      {GROUPS.map(g => (
        <div key={g.id} style={S.card}>
          <div style={S.cardHead}>
            <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <span style={S.cardTitle}>{g.name}</span>
              <ReviewBadge state={g.review_state} />
            </div>
            <span style={S.pill}>{g.member_count} members</span>
          </div>
          <div style={S.cardBody}>
            <div style={{ fontSize: 12, color: palette.textMuted, marginBottom: 12, lineHeight: 1.5 }}>{g.description}</div>
            <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
              {g.members.map((m, j) => (
                <TargetLink key={j} type={m.target_type} id={m.target_id} onClick={() => {
                  if (m.target_type === "sender") navigate("sender-detail", { email: m.target_id });
                }} />
              ))}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

function SqlWorkbenchView() {
  const [query, setQuery] = useState(PRESET_QUERIES[0].sql);
  const [activePreset, setActivePreset] = useState(0);
  const [showResults, setShowResults] = useState(false);

  const mockResults = [
    { email: "noreply@github.com", domain: "github.com", msg_count: 2341, annotations: 2 },
    { email: "newsletters@medium.com", domain: "medium.com", msg_count: 156, annotations: 1 },
    { email: "weekly@changelog.com", domain: "changelog.com", msg_count: 102, annotations: 1 },
    { email: "team@linear.app", domain: "linear.app", msg_count: 89, annotations: 1 },
    { email: "deals@shop.example.com", domain: "shop.example.com", msg_count: 47, annotations: 1 },
    { email: "hello@buttondown.email", domain: "buttondown.email", msg_count: 38, annotations: 1 },
    { email: "no-reply@accounts.google.com", domain: "accounts.google.com", msg_count: 24, annotations: 1 },
  ];

  return (
    <div style={{ display: "flex", gap: 12, height: "100%" }}>
      {/* Sidebar: presets & saved */}
      <div style={{ width: 220, flexShrink: 0 }}>
        <div style={S.card}>
          <div style={S.cardHead}><span style={S.cardTitle}>Preset Queries</span></div>
          <div>
            {PRESET_QUERIES.map((p, i) => (
              <div key={i} style={{
                padding: "8px 12px", cursor: "pointer", borderBottom: `1px solid ${palette.border}22`,
                background: activePreset === i ? palette.bgActive : "transparent",
              }} onClick={() => { setActivePreset(i); setQuery(p.sql); setShowResults(false); }}>
                <div style={{ fontSize: 11, fontFamily: font.mono, color: palette.accent }}>{p.name}</div>
                <div style={{ fontSize: 10, color: palette.textDim, marginTop: 2 }}>{p.description}</div>
              </div>
            ))}
          </div>
        </div>
        <div style={{ ...S.card, marginTop: 12 }}>
          <div style={S.cardHead}>
            <span style={S.cardTitle}>Saved Queries</span>
            <Btn>+ New</Btn>
          </div>
          <div style={{ ...S.emptyState, padding: 16, fontSize: 11 }}>No saved queries yet</div>
        </div>
      </div>

      {/* Editor & results */}
      <div style={{ flex: 1, display: "flex", flexDirection: "column", gap: 12 }}>
        <div style={{ ...S.card, flex: showResults ? "0 0 auto" : 1 }}>
          <div style={{ ...S.cardHead }}>
            <span style={S.cardTitle}>Query Editor</span>
            <div style={{ display: "flex", gap: 6 }}>
              <Btn variant="primary" onClick={() => setShowResults(true)}>▶ Run</Btn>
              <Btn>Export CSV</Btn>
            </div>
          </div>
          <div style={{ padding: 0 }}>
            <textarea
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              style={{
                width: "100%", minHeight: 160, padding: 14, background: palette.bg, color: palette.accent,
                fontFamily: font.mono, fontSize: 12, border: "none", outline: "none", resize: "vertical",
                lineHeight: 1.6, boxSizing: "border-box",
              }}
              spellCheck={false}
            />
          </div>
        </div>

        {showResults && (
          <div style={{ ...S.card, flex: 1, overflow: "auto" }}>
            <div style={S.cardHead}>
              <span style={S.cardTitle}>Results</span>
              <span style={{ fontSize: 10, color: palette.textDim, fontFamily: font.mono }}>7 rows · 2.3ms</span>
            </div>
            <table style={S.table}>
              <thead>
                <tr>
                  {Object.keys(mockResults[0]).map(k => <th key={k} style={S.th}>{k}</th>)}
                </tr>
              </thead>
              <tbody>
                {mockResults.map((row, i) => (
                  <tr key={i}>
                    {Object.entries(row).map(([k, v], j) => (
                      <td key={j} style={{ ...S.td, fontFamily: font.mono, fontSize: 12,
                        color: k === "email" ? palette.blue : typeof v === "number" ? palette.textBright : palette.text,
                        cursor: k === "email" ? "pointer" : "default",
                      }}>{typeof v === "number" ? v.toLocaleString() : v}</td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

function DashboardView({ annotations, navigate }) {
  const pending = annotations.filter(a => a.review_state === "to_review").length;
  const reviewed = annotations.filter(a => a.review_state === "reviewed").length;
  const dismissed = annotations.filter(a => a.review_state === "dismissed").length;
  const latestRun = AGENT_RUNS[0];

  return (
    <div>
      {/* Stats */}
      <div style={{ display: "flex", gap: 12, marginBottom: 16 }}>
        {[
          { label: "Pending Review", value: pending, color: palette.accent },
          { label: "Approved", value: reviewed, color: palette.green },
          { label: "Dismissed", value: dismissed, color: palette.red },
          { label: "Total Annotations", value: annotations.length, color: palette.textBright },
          { label: "Agent Runs", value: AGENT_RUNS.length, color: palette.blue },
          { label: "Senders Tracked", value: SENDERS.length, color: palette.purple },
        ].map((s, i) => (
          <div key={i} style={{ flex: 1, ...S.card, padding: 14, marginBottom: 0, textAlign: "center" }}>
            <div style={{ fontSize: 24, fontWeight: 700, fontFamily: font.mono, color: s.color }}>{s.value}</div>
            <div style={{ fontSize: 10, color: palette.textDim, textTransform: "uppercase", letterSpacing: "0.08em", fontFamily: font.mono, marginTop: 4 }}>{s.label}</div>
          </div>
        ))}
      </div>

      {/* Latest run banner */}
      {pending > 0 && (
        <div style={{ ...S.card, padding: 16, marginBottom: 16, borderColor: palette.accentDim, background: `${palette.accentDim}22` }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <div>
              <div style={{ fontSize: 13, fontWeight: 600, color: palette.accent }}>
                Latest run: <span style={S.code}>{latestRun.id}</span> ({latestRun.label})
              </div>
              <div style={{ fontSize: 12, color: palette.textMuted, marginTop: 4 }}>
                {latestRun.annotation_count} annotations produced {fmt.relative(latestRun.created_at)} · {pending} awaiting review
              </div>
            </div>
            <div style={{ display: "flex", gap: 8 }}>
              <Btn variant="primary" onClick={() => navigate("review-queue")}>Open Review Queue →</Btn>
              <Btn onClick={() => navigate("run-detail", { runId: latestRun.id })}>Inspect Run</Btn>
            </div>
          </div>
        </div>
      )}

      {/* Recent logs */}
      <div style={S.card}>
        <div style={S.cardHead}><span style={S.cardTitle}>Recent Agent Activity</span></div>
        <div style={{ padding: "4px 14px" }}>
          {LOGS.slice(0, 4).map((log, i) => (
            <div key={log.id} style={{ display: "flex", gap: 12, padding: "10px 0", borderBottom: i < 3 ? `1px solid ${palette.border}22` : "none" }}>
              <div style={{ width: 70, flexShrink: 0 }}>
                <div style={{ fontSize: 10, color: palette.textDim, fontFamily: font.mono }}>{fmt.relative(log.created_at)}</div>
              </div>
              <div>
                <span style={{ fontSize: 12, fontWeight: 600, color: palette.textBright }}>{log.title}</span>
                <span style={{ fontSize: 11, color: palette.textDim, marginLeft: 8 }}>{log.body.slice(0, 80)}…</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── App Shell ───────────────────────────────────────────────────────────────

const NAV = [
  { section: "Overview" },
  { id: "dashboard", label: "Dashboard", icon: "◫" },
  { section: "Review" },
  { id: "review-queue", label: "Review Queue", badge: true },
  { id: "agent-runs", label: "Agent Runs", icon: "⟳" },
  { id: "groups", label: "Target Groups", icon: "◫" },
  { section: "Browse" },
  { id: "senders", label: "Senders", icon: "◉" },
  { section: "Tools" },
  { id: "sql", label: "SQL Workbench", icon: "⌘" },
];

export default function App() {
  const [view, setView] = useState("dashboard");
  const [viewParams, setViewParams] = useState({});
  const [annotations, setAnnotations] = useState(ANNOTATIONS);

  const pendingCount = annotations.filter(a => a.review_state === "to_review").length;

  const navigate = useCallback((v, params = {}) => {
    setView(v);
    setViewParams(params);
  }, []);

  const topTitles = {
    dashboard: "Dashboard",
    "review-queue": "Review Queue",
    "agent-runs": "Agent Runs",
    "run-detail": `Run: ${viewParams.runId || ""}`,
    senders: "Senders",
    "sender-detail": viewParams.email || "Sender",
    groups: "Target Groups",
    sql: "SQL Workbench",
  };

  const renderView = () => {
    switch (view) {
      case "dashboard": return <DashboardView annotations={annotations} navigate={navigate} />;
      case "review-queue": return <ReviewQueueView annotations={annotations} setAnnotations={setAnnotations} navigate={navigate} />;
      case "agent-runs": return <AgentRunsView navigate={navigate} />;
      case "run-detail": return <RunDetailView runId={viewParams.runId} annotations={annotations} setAnnotations={setAnnotations} navigate={navigate} />;
      case "senders": return <SendersView annotations={annotations} navigate={navigate} />;
      case "sender-detail": return <SenderDetailView email={viewParams.email} annotations={annotations} setAnnotations={setAnnotations} navigate={navigate} />;
      case "groups": return <GroupsView navigate={navigate} />;
      case "sql": return <SqlWorkbenchView />;
      default: return <div style={S.emptyState}>View not found</div>;
    }
  };

  const breadcrumb = useMemo(() => {
    if (view === "run-detail") return [{ label: "Agent Runs", view: "agent-runs" }, { label: viewParams.runId }];
    if (view === "sender-detail") return [{ label: "Senders", view: "senders" }, { label: viewParams.email }];
    return null;
  }, [view, viewParams]);

  return (
    <div style={S.app}>
      <link href="https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700&display=swap" rel="stylesheet" />

      {/* Sidebar */}
      <nav style={S.sidebar}>
        <div style={S.sidebarHead}>
          <div style={S.sidebarLogo}>smailnail</div>
          <div style={S.sidebarSub}>annotation review</div>
        </div>
        <div style={{ flex: 1, overflowY: "auto", paddingTop: 4 }}>
          {NAV.map((item, i) =>
            item.section ? (
              <div key={i} style={S.navSection}>{item.section}</div>
            ) : (
              <div key={item.id} style={S.navItem(view === item.id || (item.id === "agent-runs" && view === "run-detail") || (item.id === "senders" && view === "sender-detail"))}
                onClick={() => navigate(item.id)}>
                <span style={{ opacity: 0.5, fontSize: 14 }}>{item.icon || "▪"}</span>
                <span>{item.label}</span>
                {item.badge && pendingCount > 0 && <span style={S.navBadge}>{pendingCount}</span>}
              </div>
            )
          )}
        </div>
        <div style={{ padding: 12, borderTop: `1px solid ${palette.border}`, fontSize: 10, color: palette.textDim, fontFamily: font.mono }}>
          3 accounts · 102,847 msgs
        </div>
      </nav>

      {/* Main */}
      <main style={S.main}>
        <header style={S.topBar}>
          {breadcrumb ? (
            <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
              {breadcrumb.map((b, i) => (
                <span key={i}>
                  {b.view ? (
                    <span style={{ ...S.link, fontSize: 14 }} onClick={() => navigate(b.view)}>{b.label}</span>
                  ) : (
                    <span style={{ ...S.topTitle }}>{b.label}</span>
                  )}
                  {i < breadcrumb.length - 1 && <span style={{ color: palette.textDim, margin: "0 4px" }}>/</span>}
                </span>
              ))}
            </div>
          ) : (
            <span style={S.topTitle}>{topTitles[view]}</span>
          )}
        </header>
        <div style={S.content}>
          {renderView()}
        </div>
      </main>
    </div>
  );
}
