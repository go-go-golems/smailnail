import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Paper from "@mui/material/Paper";
import Button from "@mui/material/Button";
import Chip from "@mui/material/Chip";
import VisibilityIcon from "@mui/icons-material/Visibility";
import { useNavigate } from "react-router-dom";
import { useListRunsQuery } from "../api/annotations";
import { SourceBadge, ReviewProgressBar } from "../components/shared";

export function AgentRunsPage() {
  const navigate = useNavigate();
  const { data: runs = [], isLoading } = useListRunsQuery();

  if (isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">
          Loading runs…
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="agent-runs-page" sx={{ p: 3 }}>
      <Typography variant="h2" sx={{ mb: 2 }}>
        Agent Runs
      </Typography>

      {runs.length === 0 ? (
        <Box sx={{ textAlign: "center", py: 6, color: "text.secondary" }}>
          <Typography variant="h4" sx={{ mb: 1 }}>
            No agent runs
          </Typography>
          <Typography variant="body2">
            Runs will appear here when agents create annotations.
          </Typography>
        </Box>
      ) : (
        <TableContainer component={Paper} sx={{ border: "none" }}>
          <Table size="small" stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>Run ID</TableCell>
                <TableCell>Source</TableCell>
                <TableCell align="right">Annotations</TableCell>
                <TableCell align="right">Logs</TableCell>
                <TableCell align="right">Groups</TableCell>
                <TableCell sx={{ minWidth: 160 }}>Progress</TableCell>
                <TableCell>Started</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {runs.map((run) => {
                const started = new Date(run.startedAt);
                const dateStr = started.toLocaleDateString("en-CA");
                const timeStr = started.toLocaleTimeString("en-GB", {
                  hour: "2-digit",
                  minute: "2-digit",
                });

                return (
                  <TableRow
                    key={run.runId}
                    hover
                    sx={{ cursor: "pointer" }}
                    onClick={() =>
                      navigate(
                        `/annotations/runs/${encodeURIComponent(run.runId)}`,
                      )
                    }
                  >
                    <TableCell>
                      <Typography
                        variant="body2"
                        sx={{ fontFamily: "monospace", fontWeight: 600 }}
                      >
                        {run.runId}
                      </Typography>
                    </TableCell>

                    <TableCell>
                      <SourceBadge
                        sourceKind={run.sourceKind}
                        sourceLabel={run.sourceLabel}
                      />
                    </TableCell>

                    <TableCell align="right">
                      <Chip
                        label={run.annotationCount}
                        size="small"
                        variant="outlined"
                        sx={{ fontFamily: "monospace", fontWeight: 700 }}
                      />
                    </TableCell>

                    <TableCell align="right">
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: "monospace",
                          color: "text.secondary",
                        }}
                      >
                        {run.logCount}
                      </Typography>
                    </TableCell>

                    <TableCell align="right">
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: "monospace",
                          color: "text.secondary",
                        }}
                      >
                        {run.groupCount}
                      </Typography>
                    </TableCell>

                    <TableCell>
                      <ReviewProgressBar
                        reviewed={run.reviewedCount}
                        pending={run.pendingCount}
                        dismissed={run.dismissedCount}
                      />
                    </TableCell>

                    <TableCell>
                      <Typography
                        variant="caption"
                        sx={{
                          fontFamily: "monospace",
                          color: "text.secondary",
                        }}
                      >
                        {dateStr} {timeStr}
                      </Typography>
                    </TableCell>

                    <TableCell onClick={(e) => e.stopPropagation()}>
                      <Button
                        size="small"
                        variant="outlined"
                        startIcon={<VisibilityIcon />}
                        onClick={() =>
                          navigate(
                            `/annotations/runs/${encodeURIComponent(run.runId)}`,
                          )
                        }
                      >
                        Inspect
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
}
