import { useMemo, useCallback, useState } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import Divider from "@mui/material/Divider";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import { useParams, useNavigate } from "react-router-dom";
import {
  useGetRunQuery,
  useBatchReviewMutation,
  useReviewAnnotationMutation,
} from "../api/annotations";
import { StatBox } from "../components/shared";
import { AnnotationTable } from "../components/AnnotationTable";
import { RunTimeline } from "../components/RunTimeline";
import { GroupCard } from "../components/GroupCard";
import type { Annotation } from "../types/annotations";

export function RunDetailPage() {
  const { runId } = useParams<{ runId: string }>();
  const navigate = useNavigate();
  const { data: run, isLoading } = useGetRunQuery(runId ?? "");
  const [batchReview] = useBatchReviewMutation();
  const [reviewAnnotation] = useReviewAnnotationMutation();

  const [selected, setSelected] = useState<string[]>([]);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const annotations = run?.annotations ?? [];
  const logs = run?.logs ?? [];
  const groups = run?.groups ?? [];

  const getRelated = useCallback(
    (ann: Annotation) =>
      annotations.filter(
        (a) =>
          a.targetType === ann.targetType &&
          a.targetId === ann.targetId &&
          a.id !== ann.id,
      ),
    [annotations],
  );

  const handleToggleSelect = useCallback((id: string) => {
    setSelected((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  }, []);

  const handleToggleAll = useCallback(() => {
    setSelected((prev) =>
      prev.length === annotations.length ? [] : annotations.map((a) => a.id),
    );
  }, [annotations]);

  const pendingIds = useMemo(
    () =>
      annotations
        .filter((a) => a.reviewState === "to_review")
        .map((a) => a.id),
    [annotations],
  );

  if (isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">
          Loading run…
        </Typography>
      </Box>
    );
  }

  if (!run) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="h4" color="error.main">
          Run not found: {runId}
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="run-detail-page" sx={{ p: 3 }}>
      {/* Header */}
      <Stack direction="row" alignItems="center" spacing={2} sx={{ mb: 2 }}>
        <Button
          size="small"
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate("/annotations/runs")}
        >
          All Runs
        </Button>
        <Typography variant="h2" sx={{ flex: 1 }}>
          Run: {run.runId}
        </Typography>
        {pendingIds.length > 0 && (
          <Button
            variant="contained"
            color="success"
            startIcon={<CheckCircleIcon />}
            onClick={() => {
              void batchReview({ ids: pendingIds, reviewState: "reviewed" });
            }}
          >
            Approve All ({pendingIds.length})
          </Button>
        )}
      </Stack>

      {/* Stat boxes */}
      <Stack
        direction="row"
        spacing={0}
        divider={<Divider orientation="vertical" flexItem />}
        sx={{
          mb: 3,
          border: 1,
          borderColor: "divider",
          borderRadius: 1,
          bgcolor: "background.paper",
        }}
      >
        <StatBox
          value={run.annotationCount}
          label="Total"
          color="text.primary"
        />
        <StatBox
          value={run.pendingCount}
          label="Pending"
          color="warning.main"
        />
        <StatBox
          value={run.reviewedCount}
          label="Reviewed"
          color="success.main"
        />
        <StatBox
          value={run.dismissedCount}
          label="Dismissed"
          color="text.secondary"
        />
      </Stack>

      {/* Timeline */}
      {logs.length > 0 && (
        <>
          <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
            Agent Log ({logs.length} entries)
          </Typography>
          <RunTimeline logs={logs} />
          <Divider sx={{ my: 3 }} />
        </>
      )}

      {/* Groups */}
      {groups.length > 0 && (
        <>
          <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
            Target Groups ({groups.length})
          </Typography>
          {groups.map((g) => (
            <GroupCard
              key={g.id}
              group={g}
              onNavigateTarget={(targetType, targetId) => {
                if (targetType === "sender") {
                  navigate(
                    `/annotations/senders/${encodeURIComponent(targetId)}`,
                  );
                }
              }}
            />
          ))}
          <Divider sx={{ my: 3 }} />
        </>
      )}

      {/* Annotations */}
      <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
        Annotations ({annotations.length})
      </Typography>
      <AnnotationTable
        annotations={annotations}
        selected={selected}
        expandedId={expandedId}
        onToggleSelect={handleToggleSelect}
        onToggleAll={handleToggleAll}
        onToggleExpand={(id) =>
          setExpandedId((prev) => (prev === id ? null : id))
        }
        onApprove={(id) =>
          void reviewAnnotation({ id, reviewState: "reviewed" })
        }
        onDismiss={(id) =>
          void reviewAnnotation({ id, reviewState: "dismissed" })
        }
        onNavigateTarget={(targetType, targetId) => {
          if (targetType === "sender") {
            navigate(
              `/annotations/senders/${encodeURIComponent(targetId)}`,
            );
          }
        }}
        getRelated={getRelated}
      />
    </Box>
  );
}
