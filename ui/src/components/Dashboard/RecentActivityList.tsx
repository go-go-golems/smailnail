import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Chip from "@mui/material/Chip";
import type { AnnotationLog } from "../../types/annotations";

export interface RecentActivityListProps {
  logs: AnnotationLog[];
  maxItems?: number;
}

export function RecentActivityList({
  logs,
  maxItems = 8,
}: RecentActivityListProps) {
  const items = logs.slice(0, maxItems);

  if (items.length === 0) {
    return (
      <Box
        data-part="recent-activity"
        data-state="empty"
        sx={{ textAlign: "center", py: 3, color: "text.secondary" }}
      >
        <Typography variant="body2">No recent activity.</Typography>
      </Box>
    );
  }

  return (
    <Paper data-part="recent-activity">
      {items.map((log, i) => {
        const date = new Date(log.createdAt);
        const timeStr = date.toLocaleTimeString("en-GB", {
          hour: "2-digit",
          minute: "2-digit",
        });
        const dateStr = date.toLocaleDateString("en-CA");

        return (
          <Box
            key={log.id}
            sx={{
              display: "flex",
              alignItems: "flex-start",
              gap: 1.5,
              px: 1.5,
              py: 1,
              borderBottom: i < items.length - 1 ? 1 : 0,
              borderColor: "divider",
            }}
          >
            <Typography
              variant="caption"
              sx={{
                fontFamily: "monospace",
                color: "text.secondary",
                minWidth: 90,
                pt: 0.25,
                flexShrink: 0,
              }}
            >
              {dateStr} {timeStr}
            </Typography>
            <Chip
              label={log.logKind}
              size="small"
              variant="outlined"
              sx={{ fontWeight: 600, fontSize: "0.625rem", flexShrink: 0 }}
            />
            <Stack sx={{ flex: 1, minWidth: 0 }}>
              <Typography variant="body2" sx={{ fontWeight: 600 }}>
                {log.title}
              </Typography>
              <Typography
                variant="caption"
                color="text.secondary"
                sx={{
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {log.bodyMarkdown.slice(0, 120)}
              </Typography>
            </Stack>
          </Box>
        );
      })}
    </Paper>
  );
}
