import { useMemo, useCallback, useState } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Stack from "@mui/material/Stack";
import { useNavigate } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "../store";
import {
  toggleSelected,
  setSelected,
  clearSelected,
  setFilterTag,
  setExpandedId,
} from "../store/annotationUiSlice";
import {
  useListAnnotationsQuery,
  useBatchReviewMutation,
  useReviewAnnotationMutation,
} from "../api/annotations";
import { AnnotationTable } from "../components/AnnotationTable";
import {
  FilterPillBar,
  CountSummaryBar,
  BatchActionBar,
} from "../components/shared";
import { ReviewCommentDrawer } from "../components/ReviewFeedback";
import type { Annotation } from "../types/annotations";
import type { FeedbackKind } from "../types/reviewFeedback";

export function ReviewQueuePage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const { selected, filterTag, expandedId } = useAppSelector(
    (s) => s.annotationUi.reviewQueue,
  );

  const { data: annotations = [], isLoading } = useListAnnotationsQuery(
    filterTag ? { tag: filterTag } : {},
  );
  const [batchReview] = useBatchReviewMutation();
  const [reviewAnnotation] = useReviewAnnotationMutation();
  const [commentDrawerOpen, setCommentDrawerOpen] = useState(false);

  // Compute tag counts for filter pills (always from unfiltered set)
  const { data: allAnnotations = [] } = useListAnnotationsQuery({});
  const tagCounts = useMemo(() => {
    const counts = new Map<string, number>();
    for (const ann of allAnnotations) {
      counts.set(ann.tag, (counts.get(ann.tag) ?? 0) + 1);
    }
    return Array.from(counts.entries())
      .map(([key, count]) => ({ key, label: key, count }))
      .sort((a, b) => b.count - a.count);
  }, [allAnnotations]);

  // Summary counts for current view
  const summaryItems = useMemo(() => {
    const toReview = annotations.filter(
      (a) => a.reviewState === "to_review",
    ).length;
    const agentCount = annotations.filter(
      (a) => a.sourceKind === "agent",
    ).length;
    const heuristicCount = annotations.filter(
      (a) => a.sourceKind === "heuristic",
    ).length;
    return [
      { label: "to review", value: toReview, color: "#d29922" },
      { label: "agent", value: agentCount, color: "#58a6ff" },
      { label: "heuristic", value: heuristicCount },
    ];
  }, [annotations]);

  const handleBatchRejectExplain = useCallback(() => {
    setCommentDrawerOpen(true);
  }, []);

  const handleCommentSubmit = useCallback(
    (payload: { feedbackKind: FeedbackKind; title: string; bodyMarkdown: string; guidelineIds: string[] }) => {
      void batchReview({
        ids: selected,
        reviewState: "dismissed",
        comment: {
          feedbackKind: payload.feedbackKind,
          title: payload.title,
          bodyMarkdown: payload.bodyMarkdown,
        },
        guidelineIds: payload.guidelineIds.length > 0 ? payload.guidelineIds : undefined,
      });
      dispatch(clearSelected());
      setCommentDrawerOpen(false);
    },
    [batchReview, selected, dispatch],
  );

  const handleGetRelated = useCallback(
    (ann: Annotation) =>
      annotations.filter(
        (a) =>
          a.targetType === ann.targetType &&
          a.targetId === ann.targetId &&
          a.id !== ann.id,
      ),
    [annotations],
  );

  const handleToggleAll = useCallback(() => {
    if (selected.length === annotations.length) {
      dispatch(clearSelected());
    } else {
      dispatch(setSelected(annotations.map((a) => a.id)));
    }
  }, [dispatch, selected.length, annotations]);

  const handleBatchApprove = useCallback(() => {
    void batchReview({ ids: selected, reviewState: "reviewed" });
    dispatch(clearSelected());
  }, [batchReview, selected, dispatch]);

  const handleBatchDismiss = useCallback(() => {
    void batchReview({ ids: selected, reviewState: "dismissed" });
    dispatch(clearSelected());
  }, [batchReview, selected, dispatch]);

  const handleBatchReset = useCallback(() => {
    void batchReview({ ids: selected, reviewState: "to_review" });
    dispatch(clearSelected());
  }, [batchReview, selected, dispatch]);

  const handleNavigateTarget = useCallback(
    (targetType: string, targetId: string) => {
      if (targetType === "sender") {
        navigate(`/annotations/senders/${encodeURIComponent(targetId)}`);
      }
    },
    [navigate],
  );

  if (isLoading) {
    return (
      <Box sx={{ p: 3 }}>
        <Typography variant="body2" color="text.secondary">
          Loading annotations…
        </Typography>
      </Box>
    );
  }

  return (
    <Box data-widget="review-queue-page" sx={{ p: 3 }}>
      <Typography variant="h2" sx={{ mb: 2 }}>
        Review Queue
      </Typography>

      <Stack spacing={1.5} sx={{ mb: 2 }}>
        <FilterPillBar
          pills={tagCounts}
          activeKey={filterTag}
          onSelect={(key) => dispatch(setFilterTag(key))}
        />
        <CountSummaryBar items={summaryItems} />
      </Stack>

      <BatchActionBar
        totalCount={annotations.length}
        selectedCount={selected.length}
        allSelected={
          annotations.length > 0 && selected.length === annotations.length
        }
        onToggleAll={handleToggleAll}
        onApprove={handleBatchApprove}
        onDismiss={handleBatchDismiss}
        onRejectExplain={handleBatchRejectExplain}
        onReset={handleBatchReset}
      />

      <AnnotationTable
        annotations={annotations}
        selected={selected}
        expandedId={expandedId}
        onToggleSelect={(id) => dispatch(toggleSelected(id))}
        onToggleAll={handleToggleAll}
        onToggleExpand={(id) =>
          dispatch(setExpandedId(expandedId === id ? null : id))
        }
        onApprove={(id) =>
          void reviewAnnotation({ id, reviewState: "reviewed" })
        }
        onDismiss={(id) =>
          void reviewAnnotation({ id, reviewState: "dismissed" })
        }
        onNavigateTarget={handleNavigateTarget}
        getRelated={handleGetRelated}
      />

      <ReviewCommentDrawer
        open={commentDrawerOpen}
        mode="batch"
        targetCount={selected.length}
        onSubmit={handleCommentSubmit}
        onCancel={() => setCommentDrawerOpen(false)}
      />
    </Box>
  );
}
