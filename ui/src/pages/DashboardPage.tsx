import { useMemo } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import { useNavigate } from "react-router-dom";
import {
  useListAnnotationsQuery,
  useListRunsQuery,
  useListLogsQuery,
  useListSendersQuery,
} from "../api/annotations";
import {
  DashboardStatGrid,
  LatestRunBanner,
  RecentActivityList,
} from "../components/Dashboard";

export function DashboardPage() {
  const navigate = useNavigate();
  const { data: annotations = [] } = useListAnnotationsQuery({});
  const { data: runs = [] } = useListRunsQuery();
  const { data: logs = [] } = useListLogsQuery({});
  const { data: senders = [] } = useListSendersQuery({});

  const stats = useMemo(() => {
    const toReview = annotations.filter(
      (a) => a.reviewState === "to_review",
    ).length;
    const reviewed = annotations.filter(
      (a) => a.reviewState === "reviewed",
    ).length;
    const dismissed = annotations.filter(
      (a) => a.reviewState === "dismissed",
    ).length;

    return [
      { value: toReview, label: "To Review", color: "#d29922" },
      { value: reviewed, label: "Reviewed", color: "#3fb950" },
      { value: dismissed, label: "Dismissed" },
      { value: annotations.length, label: "Total" },
      { value: runs.length, label: "Agent Runs", color: "#58a6ff" },
      { value: senders.length, label: "Senders", color: "#a78bfa" },
    ];
  }, [annotations, runs.length, senders.length]);

  // Latest run with pending annotations
  const latestRun = useMemo(
    () =>
      [...runs]
        .sort(
          (a, b) =>
            new Date(b.startedAt).getTime() -
            new Date(a.startedAt).getTime(),
        )
        .find((r) => r.pendingCount > 0) ?? runs[0],
    [runs],
  );

  // Recent logs sorted by date
  const recentLogs = useMemo(
    () =>
      [...logs].sort(
        (a, b) =>
          new Date(b.createdAt).getTime() -
          new Date(a.createdAt).getTime(),
      ),
    [logs],
  );

  return (
    <Box data-widget="dashboard-page" sx={{ p: 3 }}>
      <Typography variant="h2" sx={{ mb: 3 }}>
        Annotations Dashboard
      </Typography>

      <Stack spacing={3}>
        <DashboardStatGrid stats={stats} />

        {latestRun && (
          <LatestRunBanner
            run={latestRun}
            onReview={() => navigate("/annotations/review")}
            onInspect={() =>
              navigate(
                `/annotations/runs/${encodeURIComponent(latestRun.runId)}`,
              )
            }
          />
        )}

        <Box>
          <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
            Recent Agent Activity
          </Typography>
          <RecentActivityList logs={recentLogs} />
        </Box>
      </Stack>
    </Box>
  );
}
