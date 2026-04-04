import Chip from "@mui/material/Chip";
import type { ReviewState } from "../../types/annotations";
import { parts } from "./parts";

export interface ReviewStateBadgeProps {
  state: ReviewState;
  size?: "small" | "medium";
}

const stateConfig: Record<
  ReviewState,
  { label: string; color: "warning" | "success" | "default" }
> = {
  to_review: { label: "To Review", color: "warning" },
  reviewed: { label: "Reviewed", color: "success" },
  dismissed: { label: "Dismissed", color: "default" },
};

export function ReviewStateBadge({
  state,
  size = "small",
}: ReviewStateBadgeProps) {
  const config = stateConfig[state];

  return (
    <Chip
      data-part={parts.reviewBadge}
      data-state={state}
      label={config.label}
      size={size}
      color={config.color}
      variant="outlined"
      sx={{ fontWeight: 600, fontSize: "0.6875rem" }}
    />
  );
}
