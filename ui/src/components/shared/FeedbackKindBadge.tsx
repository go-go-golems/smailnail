import Chip from "@mui/material/Chip";
import type { FeedbackKind } from "../../types/reviewFeedback";
import { parts } from "./parts";

export interface FeedbackKindBadgeProps {
  kind: FeedbackKind;
  size?: "small" | "medium";
}

const kindConfig: Record<
  FeedbackKind,
  { label: string; color: "error" | "info" | "warning" | "default" }
> = {
  reject_request: { label: "Reject Request", color: "error" },
  comment: { label: "Comment", color: "info" },
  guideline_request: { label: "Guideline Request", color: "warning" },
  clarification: { label: "Clarification", color: "default" },
};

export function FeedbackKindBadge({
  kind,
  size = "small",
}: FeedbackKindBadgeProps) {
  const config = kindConfig[kind];

  return (
    <Chip
      data-part={parts.feedbackKindBadge}
      data-state={kind}
      label={config.label}
      size={size}
      color={config.color}
      variant="outlined"
      sx={{ fontWeight: 600, fontSize: "0.6875rem" }}
    />
  );
}
