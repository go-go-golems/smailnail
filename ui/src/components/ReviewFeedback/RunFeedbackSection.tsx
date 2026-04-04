import { useState } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import AddCommentIcon from "@mui/icons-material/AddComment";
import {
  useCreateReviewFeedbackMutation,
  useUpdateReviewFeedbackMutation,
} from "../../api/annotations";
import { FeedbackCard } from "./FeedbackCard";
import { ReviewCommentDrawer } from "./ReviewCommentDrawer";
import type { ReviewFeedback, FeedbackKind } from "../../types/reviewFeedback";
import { parts } from "./parts";

export interface RunFeedbackSectionProps {
  runId: string;
  feedback: ReviewFeedback[];
}

export function RunFeedbackSection({
  runId,
  feedback,
}: RunFeedbackSectionProps) {
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [createFeedback] = useCreateReviewFeedbackMutation();
  const [updateFeedback] = useUpdateReviewFeedbackMutation();

  const handleSubmit = (payload: {
    feedbackKind: FeedbackKind;
    title: string;
    bodyMarkdown: string;
    guidelineIds: string[];
  }) => {
    createFeedback({
      scopeKind: "run",
      agentRunId: runId,
      feedbackKind: payload.feedbackKind,
      title: payload.title,
      bodyMarkdown: payload.bodyMarkdown,
    });
    setDrawerOpen(false);
  };

  return (
    <Box data-part={parts.feedbackPanel}>
      <Typography variant="overline" sx={{ display: "block", mb: 1.5 }}>
        Run-Level Feedback ({feedback.length})
      </Typography>

      {feedback.length === 0 && !drawerOpen && (
        <Box
          sx={{
            textAlign: "center",
            py: 2,
            color: "text.secondary",
            border: 1,
            borderColor: "divider",
            borderRadius: 1,
            mb: 1,
          }}
        >
          <Typography variant="body2">No feedback yet.</Typography>
        </Box>
      )}

      {feedback.map((fb) => (
        <FeedbackCard
          key={fb.id}
          feedback={fb}
          onAcknowledge={
            fb.status === "open"
              ? () => updateFeedback({ id: fb.id, status: "acknowledged" })
              : undefined
          }
          onResolve={
            fb.status !== "resolved"
              ? () => updateFeedback({ id: fb.id, status: "resolved" })
              : undefined
          }
        />
      ))}

      {!drawerOpen && (
        <Button
          size="small"
          variant="text"
          startIcon={<AddCommentIcon />}
          onClick={() => setDrawerOpen(true)}
          sx={{ mt: 1 }}
        >
          Add Run Feedback
        </Button>
      )}

      <ReviewCommentDrawer
        open={drawerOpen}
        mode="run"
        targetCount={0}
        agentRunId={runId}
        onSubmit={handleSubmit}
        onCancel={() => setDrawerOpen(false)}
      />
    </Box>
  );
}
