import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import type { SenderGuidelineGroup } from "../../types/annotations";
import { GuidelineCard } from "../RunGuideline/GuidelineCard";

export interface SenderGuidelinePanelProps {
  groups: SenderGuidelineGroup[];
  onNavigateRun?: (runId: string) => void;
}

export function SenderGuidelinePanel({
  groups,
  onNavigateRun,
}: SenderGuidelinePanelProps) {
  if (groups.length === 0) {
    return null;
  }

  return (
    <Box data-widget="sender-guideline-panel" sx={{ mb: 3 }}>
      <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
        Linked Guidelines ({groups.reduce((total, group) => total + group.guidelines.length, 0)})
      </Typography>

      <Stack spacing={1.5}>
        {groups.map((group) => (
          <Paper key={group.runId} variant="outlined" sx={{ p: 1.5 }}>
            <Stack
              direction={{ xs: "column", sm: "row" }}
              alignItems={{ xs: "flex-start", sm: "center" }}
              spacing={1}
              sx={{ mb: 1 }}
            >
              <Typography variant="body2" sx={{ fontWeight: 600, flex: 1 }}>
                Run: {group.runId}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {group.sourceLabel || group.sourceKind || "Unknown source"}
              </Typography>
              {onNavigateRun && (
                <Button
                  size="small"
                  variant="text"
                  startIcon={<OpenInNewIcon />}
                  onClick={() => onNavigateRun(group.runId)}
                >
                  Open run
                </Button>
              )}
            </Stack>

            {group.guidelines.map((guideline) => (
              <GuidelineCard
                key={`${group.runId}-${guideline.id}`}
                guideline={guideline}
                compact
              />
            ))}
          </Paper>
        ))}
      </Stack>
    </Box>
  );
}
