import Box from "@mui/material/Box";
import Tooltip from "@mui/material/Tooltip";
import { parts } from "./parts";

export interface ReviewProgressBarProps {
  reviewed: number;
  pending: number;
  dismissed: number;
  /** Height in pixels */
  height?: number;
}

export function ReviewProgressBar({
  reviewed,
  pending,
  dismissed,
  height = 6,
}: ReviewProgressBarProps) {
  const total = reviewed + pending + dismissed;
  if (total === 0) return null;

  const reviewedPct = (reviewed / total) * 100;
  const dismissedPct = (dismissed / total) * 100;
  const pendingPct = (pending / total) * 100;

  const tooltip = `${reviewed} reviewed · ${pending} pending · ${dismissed} dismissed`;

  return (
    <Tooltip title={tooltip} arrow>
      <Box
        data-part={parts.reviewProgress}
        sx={{
          display: "flex",
          width: "100%",
          height,
          borderRadius: height / 2,
          overflow: "hidden",
          bgcolor: "action.hover",
        }}
      >
        {reviewedPct > 0 && (
          <Box
            data-state="reviewed"
            sx={{
              width: `${reviewedPct}%`,
              bgcolor: "success.main",
              transition: "width 0.3s ease",
            }}
          />
        )}
        {dismissedPct > 0 && (
          <Box
            data-state="dismissed"
            sx={{
              width: `${dismissedPct}%`,
              bgcolor: "text.secondary",
              opacity: 0.5,
              transition: "width 0.3s ease",
            }}
          />
        )}
        {pendingPct > 0 && (
          <Box
            data-state="pending"
            sx={{
              width: `${pendingPct}%`,
              bgcolor: "warning.main",
              transition: "width 0.3s ease",
            }}
          />
        )}
      </Box>
    </Tooltip>
  );
}
