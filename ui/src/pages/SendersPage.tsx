import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Chip from "@mui/material/Chip";
import UnsubscribeIcon from "@mui/icons-material/Unsubscribe";
import { useNavigate } from "react-router-dom";
import { useListSendersQuery } from "../api/annotations";
import { TagChip } from "../components/shared";

export function SendersPage() {
  const navigate = useNavigate();
  const { data: senders = [], isLoading } = useListSendersQuery({});

  if (isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">
          Loading senders…
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="senders-page" sx={{ p: 3 }}>
      <Typography variant="h2" sx={{ mb: 2 }}>
        Senders
      </Typography>

      {senders.length === 0 ? (
        <Box sx={{ textAlign: "center", py: 6, color: "text.secondary" }}>
          <Typography variant="h4" sx={{ mb: 1 }}>
            No senders
          </Typography>
          <Typography variant="body2">
            Senders will appear here after messages are synced.
          </Typography>
        </Box>
      ) : (
        <TableContainer component={Paper} sx={{ border: "none" }}>
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>Email</TableCell>
                <TableCell>Display Name</TableCell>
                <TableCell>Domain</TableCell>
                <TableCell align="right">Messages</TableCell>
                <TableCell align="right">Annotations</TableCell>
                <TableCell>Tags</TableCell>
                <TableCell align="center">Unsub</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {senders.map((sender) => (
                <TableRow
                  key={sender.email}
                  hover
                  sx={{ cursor: "pointer" }}
                  onClick={() =>
                    navigate(
                      `/annotations/senders/${encodeURIComponent(sender.email)}`,
                    )
                  }
                >
                  <TableCell>
                    <Typography
                      variant="body2"
                      sx={{ fontFamily: "monospace", fontWeight: 600 }}
                    >
                      {sender.email}
                    </Typography>
                  </TableCell>

                  <TableCell>
                    <Typography variant="body2">
                      {sender.displayName || "—"}
                    </Typography>
                  </TableCell>

                  <TableCell>
                    <Typography
                      variant="body2"
                      sx={{ fontFamily: "monospace", color: "text.secondary" }}
                    >
                      {sender.domain}
                    </Typography>
                  </TableCell>

                  <TableCell align="right">
                    <Chip
                      label={sender.messageCount}
                      size="small"
                      variant="outlined"
                      sx={{ fontFamily: "monospace", fontWeight: 700 }}
                    />
                  </TableCell>

                  <TableCell align="right">
                    <Typography
                      variant="body2"
                      sx={{ fontFamily: "monospace", color: "text.secondary" }}
                    >
                      {sender.annotationCount}
                    </Typography>
                  </TableCell>

                  <TableCell>
                    <Stack direction="row" spacing={0.5}>
                      {sender.tags.map((tag) => (
                        <TagChip key={tag} tag={tag} />
                      ))}
                    </Stack>
                  </TableCell>

                  <TableCell align="center">
                    {sender.hasUnsubscribe && (
                      <UnsubscribeIcon
                        fontSize="small"
                        sx={{ color: "text.secondary" }}
                        titleAccess="Has unsubscribe header"
                      />
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
}
