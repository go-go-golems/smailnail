import Paper from "@mui/material/Paper";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import RateReviewIcon from "@mui/icons-material/RateReview";
import VisibilityIcon from "@mui/icons-material/Visibility";
import type { AgentRunSummary } from "../../types/annotations";
import { SourceBadge, ReviewProgressBar } from "../shared";

export interface LatestRunBannerProps {
  run: AgentRunSummary;
  onReview: () => void;
  onInspect: () => void;
}

export function LatestRunBanner({
  run,
  onReview,
  onInspect,
}: LatestRunBannerProps) {
  const started = new Date(run.startedAt);
  const timeAgo = getTimeAgo(started);

  return (
    <Paper
      data-part="latest-run-banner"
      sx={{
        p: 2,
        border: "1px solid",
        borderColor: "primary.dark",
        bgcolor: "rgba(245, 166, 35, 0.04)",
      }}
    >
      <Stack
        direction="row"
        alignItems="center"
        spacing={2}
        sx={{ mb: 1.5 }}
      >
        <Typography variant="body1" sx={{ fontWeight: 600, flex: 1 }}>
          Latest Run: {run.runId}
        </Typography>
        <SourceBadge
          sourceKind={run.sourceKind}
          sourceLabel={run.sourceLabel}
        />
        <Typography variant="caption" color="text.secondary">
          {timeAgo}
        </Typography>
      </Stack>

      <Box sx={{ mb: 1.5 }}>
        <ReviewProgressBar
          reviewed={run.reviewedCount}
          pending={run.pendingCount}
          dismissed={run.dismissedCount}
          height={8}
        />
      </Box>

      <Stack direction="row" spacing={1.5}>
        <Typography
          variant="body2"
          sx={{ flex: 1, color: "text.secondary" }}
        >
          {run.pendingCount} annotations pending review ·{" "}
          {run.annotationCount} total
        </Typography>
        {run.pendingCount > 0 && (
          <Button
            variant="contained"
            size="small"
            startIcon={<RateReviewIcon />}
            onClick={onReview}
          >
            Review Queue
          </Button>
        )}
        <Button
          variant="outlined"
          size="small"
          startIcon={<VisibilityIcon />}
          onClick={onInspect}
        >
          Inspect Run
        </Button>
      </Stack>
    </Paper>
  );
}

function getTimeAgo(date: Date): string {
  const diffMs = Date.now() - date.getTime();
  const diffMin = Math.floor(diffMs / 60_000);
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;
  const diffDays = Math.floor(diffHr / 24);
  return `${diffDays}d ago`;
}
