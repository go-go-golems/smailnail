import Chip from "@mui/material/Chip";
import type { FeedbackStatus } from "../../types/reviewFeedback";
import { parts } from "./parts";

export interface FeedbackStatusBadgeProps {
  status: FeedbackStatus;
  size?: "small" | "medium";
}

const statusConfig: Record<
  FeedbackStatus,
  { label: string; color: "warning" | "info" | "success" | "default" }
> = {
  open: { label: "Open", color: "warning" },
  acknowledged: { label: "Acknowledged", color: "info" },
  resolved: { label: "Resolved", color: "success" },
  archived: { label: "Archived", color: "default" },
};

export function FeedbackStatusBadge({
  status,
  size = "small",
}: FeedbackStatusBadgeProps) {
  const config = statusConfig[status];

  return (
    <Chip
      data-part={parts.feedbackStatusBadge}
      data-state={status}
      label={config.label}
      size={size}
      color={config.color}
      variant="outlined"
      sx={{ fontWeight: 600, fontSize: "0.6875rem" }}
    />
  );
}
