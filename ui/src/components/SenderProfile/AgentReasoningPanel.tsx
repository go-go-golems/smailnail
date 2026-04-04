import Paper from "@mui/material/Paper";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Chip from "@mui/material/Chip";
import type { AnnotationLog } from "../../types/annotations";
import { MarkdownRenderer } from "../shared";

export interface AgentReasoningPanelProps {
  logs: AnnotationLog[];
}

export function AgentReasoningPanel({ logs }: AgentReasoningPanelProps) {
  if (logs.length === 0) {
    return null;
  }

  return (
    <Box data-part="agent-reasoning">
      {logs.map((log) => {
        const date = new Date(log.createdAt);
        const dateStr = date.toLocaleDateString("en-CA");
        const timeStr = date.toLocaleTimeString("en-GB", {
          hour: "2-digit",
          minute: "2-digit",
        });

        return (
          <Paper key={log.id} sx={{ p: 1.5, mb: 1 }}>
            <Stack
              direction="row"
              spacing={1}
              alignItems="center"
              sx={{ mb: 1 }}
            >
              <Chip
                label={log.logKind}
                size="small"
                variant="outlined"
                color="info"
                sx={{ fontWeight: 600, fontSize: "0.6875rem" }}
              />
              <Typography variant="body2" sx={{ fontWeight: 600, flex: 1 }}>
                {log.title}
              </Typography>
              <Typography
                variant="caption"
                sx={{ fontFamily: "monospace", color: "text.secondary" }}
              >
                {dateStr} {timeStr}
              </Typography>
            </Stack>
            <MarkdownRenderer content={log.bodyMarkdown} />
          </Paper>
        );
      })}
    </Box>
  );
}
