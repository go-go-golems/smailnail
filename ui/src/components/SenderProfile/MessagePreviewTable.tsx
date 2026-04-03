import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import type { MessagePreview } from "../../types/annotations";

export interface MessagePreviewTableProps {
  messages: MessagePreview[];
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function MessagePreviewTable({ messages }: MessagePreviewTableProps) {
  if (messages.length === 0) {
    return (
      <Box
        data-part="message-preview"
        data-state="empty"
        sx={{ textAlign: "center", py: 3, color: "text.secondary" }}
      >
        <Typography variant="body2">No messages available.</Typography>
      </Box>
    );
  }

  return (
    <TableContainer component={Paper} data-part="message-preview" sx={{ border: "none" }}>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Date</TableCell>
            <TableCell>Subject</TableCell>
            <TableCell align="right">Size</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {messages.map((msg) => {
            const date = new Date(msg.date);
            const dateStr = date.toLocaleDateString("en-CA");

            return (
              <TableRow key={msg.uid} hover>
                <TableCell>
                  <Typography
                    variant="caption"
                    sx={{ fontFamily: "monospace", color: "text.secondary" }}
                  >
                    {dateStr}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography
                    variant="body2"
                    sx={{
                      maxWidth: 400,
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {msg.subject}
                  </Typography>
                </TableCell>
                <TableCell align="right">
                  <Typography
                    variant="caption"
                    sx={{ fontFamily: "monospace", color: "text.secondary" }}
                  >
                    {formatBytes(msg.sizeBytes)}
                  </Typography>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
