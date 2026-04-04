import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Chip from "@mui/material/Chip";
import type { AnnotationLog } from "../../types/annotations";
import { MarkdownRenderer, SourceBadge } from "../shared";

export interface RunTimelineProps {
  logs: AnnotationLog[];
  onNavigateTarget?: (targetType: string, targetId: string) => void;
}

const kindColors: Record<string, "info" | "success" | "warning" | "error" | "default"> = {
  reasoning: "info",
  decision: "warning",
  summary: "success",
  error: "error",
};

export function RunTimeline({ logs }: RunTimelineProps) {
  if (logs.length === 0) {
    return (
      <Box
        data-widget="run-timeline"
        data-state="empty"
        sx={{ textAlign: "center", py: 4, color: "text.secondary" }}
      >
        <Typography variant="body2">No log entries for this run.</Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="run-timeline">
      {logs.map((log) => {
        const date = new Date(log.createdAt);
        const timeStr = date.toLocaleTimeString("en-GB", {
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        });

        return (
          <Box
            key={log.id}
            data-part="timeline-entry"
            sx={{
              display: "flex",
              gap: 2,
              mb: 1.5,
              "&:last-child": { mb: 0 },
            }}
          >
            {/* Time column */}
            <Typography
              variant="caption"
              sx={{
                fontFamily: "monospace",
                color: "text.secondary",
                minWidth: 70,
                pt: 1.25,
                flexShrink: 0,
              }}
            >
              {timeStr}
            </Typography>

            {/* Connector line */}
            <Box
              sx={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                pt: 1.25,
                flexShrink: 0,
              }}
            >
              <Box
                sx={{
                  width: 8,
                  height: 8,
                  borderRadius: "50%",
                  bgcolor: `${kindColors[log.logKind] ?? "default"}.main`,
                  flexShrink: 0,
                }}
              />
              <Box
                sx={{
                  width: 1,
                  flex: 1,
                  bgcolor: "divider",
                  mt: 0.5,
                }}
              />
            </Box>

            {/* Content */}
            <Paper
              data-part="timeline-card"
              sx={{ flex: 1, p: 1.5, mb: 0 }}
            >
              <Stack
                direction="row"
                spacing={1}
                alignItems="center"
                sx={{ mb: 1 }}
              >
                <Chip
                  label={log.logKind}
                  size="small"
                  color={kindColors[log.logKind] ?? "default"}
                  variant="outlined"
                  sx={{ fontWeight: 600, fontSize: "0.6875rem" }}
                />
                <Typography variant="body2" sx={{ fontWeight: 600, flex: 1 }}>
                  {log.title}
                </Typography>
                <SourceBadge
                  sourceKind={log.sourceKind}
                  sourceLabel={log.sourceLabel}
                />
              </Stack>
              <MarkdownRenderer content={log.bodyMarkdown} />
            </Paper>
          </Box>
        );
      })}
    </Box>
  );
}
