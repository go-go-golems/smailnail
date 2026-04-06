import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Chip from "@mui/material/Chip";
import IconButton from "@mui/material/IconButton";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import { parts } from "./parts";
import type { AgentRunSummary } from "../../types/annotations";

export interface GuidelineLinkedRunsProps {
  runs: AgentRunSummary[];
  onNavigateRun?: (runId: string) => void;
}

export function GuidelineLinkedRuns({
  runs,
  onNavigateRun,
}: GuidelineLinkedRunsProps) {
  if (runs.length === 0) {
    return (
      <Box data-part={parts.guidelineLinkedRuns} sx={{ py: 2, textAlign: "center" }}>
        <Typography variant="body2" color="text.secondary">
          No runs linked to this guideline.
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-part={parts.guidelineLinkedRuns}>
      <Typography variant="overline" sx={{ display: "block", mb: 1 }}>
        Linked Runs ({runs.length})
      </Typography>
      <Stack spacing={0.5}>
        {runs.map((run) => (
          <Box
            key={run.runId}
            sx={{
              display: "flex",
              alignItems: "center",
              gap: 1,
              p: 1,
              borderRadius: 1,
              border: 1,
              borderColor: "divider",
              "&:hover": { bgcolor: "action.hover" },
            }}
          >
            <Typography variant="body2" sx={{ flex: 1, fontWeight: 500 }}>
              {run.sourceLabel}
            </Typography>
            <Chip
              label={`${run.annotationCount} annotations`}
              size="small"
              variant="outlined"
            />
            {run.pendingCount > 0 && (
              <Chip
                label={`${run.pendingCount} pending`}
                size="small"
                color="warning"
                variant="outlined"
              />
            )}
            <Typography variant="caption" color="text.secondary">
              {new Date(run.startedAt).toLocaleDateString()}
            </Typography>
            {onNavigateRun && (
              <IconButton
                size="small"
                onClick={() => onNavigateRun(run.runId)}
                aria-label={`Navigate to run ${run.runId}`}
              >
                <OpenInNewIcon fontSize="small" />
              </IconButton>
            )}
          </Box>
        ))}
      </Stack>
    </Box>
  );
}
