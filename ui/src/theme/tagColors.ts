/**
 * Per-tag color map for annotation tags.
 * Each tag gets a background, foreground, and border color
 * tuned for readability on the dark theme.
 */
export interface TagColorEntry {
  bg: string;
  fg: string;
  border: string;
}

export const tagColors: Record<string, TagColorEntry> = {
  newsletter: { bg: "#1e293b", fg: "#60a5fa", border: "#334155" },
  notification: { bg: "#1c1917", fg: "#fb923c", border: "#3f3224" },
  "bulk-sender": { bg: "#1a1625", fg: "#a78bfa", border: "#2e2544" },
  important: { bg: "#14231a", fg: "#4ade80", border: "#1e3a28" },
  transactional: { bg: "#1a1a14", fg: "#d4a053", border: "#33301e" },
  ignore: { bg: "#1c1115", fg: "#f87171", border: "#3b1a22" },
  "action-required": { bg: "#231414", fg: "#f87171", border: "#3b1a1a" },
  personal: { bg: "#1a2332", fg: "#93c5fd", border: "#1e3a5f" },
  marketing: { bg: "#1f1a2e", fg: "#c4b5fd", border: "#312e44" },
  spam: { bg: "#2a1414", fg: "#fca5a5", border: "#4b1a1a" },
};

/** Fallback colors for unknown tags */
export const tagColorFallback: TagColorEntry = {
  bg: "#1a1d23",
  fg: "#8b949e",
  border: "#30363d",
};

/** Get colors for a tag, falling back to neutral if unknown */
export function getTagColor(tag: string): TagColorEntry {
  return tagColors[tag] ?? tagColorFallback;
}
